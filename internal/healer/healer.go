package healer

import (
	"log"
	"os"
	"os/exec"
	"time"

	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/srvsurya/system-monitor/internal/models"
)

const (
	AnomalyMultiplier = 2.0
	CooldownMinutes   = 5
	MinSamples        = 10
)

type Healer struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Healer {
	return &Healer{db: db}
}

func (h *Healer) Evaluate(procs []models.ManagedProcess) {
	log.Printf("[healer] evaluating %d processes", len(procs))
	for _, p := range procs {
		if p.Status != "running" {
			continue
		}

		// get live cpu/mem for this process
		cpu, mem, err := getLiveStats(p.PID)
		if err != nil {
			continue // process might have died, skip
		}

		baseline := h.getOrCreate(p.Name)

		// update rolling average
		baseline = h.updateBaseline(baseline, cpu, mem)

		// not enough data yet
		if baseline.SampleCount < MinSamples {
			continue
		}

		// check cooldown
		if baseline.LastHealedAt != nil {
			lastHealed, err := time.Parse(time.RFC3339, *baseline.LastHealedAt)
			log.Printf("[healer] last healed: %v, since: %v, cooldown: %v, err: %v",
				lastHealed, time.Now().UTC().Sub(lastHealed), CooldownMinutes*time.Minute, err)
			if err == nil && time.Now().UTC().Sub(lastHealed) < CooldownMinutes*time.Minute {
				log.Printf("[healer] %s in cooldown, skipping", p.Name)
				continue
			}
		}

		// check anomaly
		cpuAnomaly := cpu > baseline.AvgCPU*AnomalyMultiplier
		memAnomaly := mem > baseline.AvgMemory*AnomalyMultiplier

		if cpuAnomaly || memAnomaly {
			var reason string
			var metric float64

			if cpuAnomaly && memAnomaly {
				reason = "cpu and memory anomaly"
				metric = cpu
			} else if cpuAnomaly {
				reason = "cpu anomaly"
				metric = cpu
			} else {
				reason = "memory anomaly"
				metric = mem
			}

			h.heal(p, reason, metric)
		}
	}
}

func getLiveStats(pid int) (cpu float64, mem float64, err error) {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return 0, 0, err
	}
	cpu, err = p.CPUPercent()
	if err != nil {
		return 0, 0, err
	}
	memInfo, err := p.MemoryInfo()
	if err != nil {
		return 0, 0, err
	}
	mem = float64(memInfo.RSS / 1024 / 1024) // MB
	return cpu, mem, nil
}

func (h *Healer) getOrCreate(name string) models.ProcessBaseline {
	var baseline models.ProcessBaseline
	err := h.db.Get(&baseline, `SELECT * FROM process_baselines WHERE process_name = ?`, name)
	if err != nil {
		// doesn't exist yet, insert a fresh one
		h.db.Exec(`INSERT INTO process_baselines (process_name) VALUES (?)`, name)
		h.db.Get(&baseline, `SELECT * FROM process_baselines WHERE process_name = ?`, name)
	}
	return baseline
}
func (h *Healer) updateBaseline(b models.ProcessBaseline, cpu float64, mem float64) models.ProcessBaseline {
	count := float64(b.SampleCount)
	b.AvgCPU = (b.AvgCPU*count + cpu) / (count + 1)
	b.AvgMemory = (b.AvgMemory*count + mem) / (count + 1)
	b.SampleCount++

	h.db.Exec(`
        UPDATE process_baselines
        SET avg_cpu = ?, avg_memory = ?, sample_count = ?, last_updated = datetime('now')
        WHERE process_name = ?`,
		b.AvgCPU, b.AvgMemory, b.SampleCount, b.ProcessName,
	)
	return b
}
func (h *Healer) heal(p models.ManagedProcess, reason string, val float64) {
	log.Printf("[healer] anomaly detected on %s (%s) — restarting", p.Name, reason)

	// kill
	proc, err := os.FindProcess(p.PID)
	if err == nil {
		proc.Kill()
	}

	// restart — same as RestartProcess handler
	if p.Command == nil || *p.Command == "" {
		log.Printf("[healer] no command stored for %s, cannot restart", p.Name)
		return
	}
	parts := strings.Fields(*p.Command)
	if len(parts) == 0 {
		log.Printf("[healer] command stored for %s not found, cannot restart", p.Name)
		return
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	if err := cmd.Start(); err != nil {
		log.Printf("[healer] failed to restart %s: %v", p.Name, err)
		return
	}

	newPID := cmd.Process.Pid

	// update managed_processes
	h.db.Exec(`UPDATE managed_processes SET pid = ?, status = 'running' WHERE id = ?`, newPID, p.ID)

	// log the action
	h.db.Exec(`
        INSERT INTO system_actions (process_id, process_name, action_type, reason, smart_heal_metric)
        VALUES (?, ?,'smart_heal',?, ?)`, p.ID, p.Name, reason, val)

	// update cooldown
	h.db.Exec(`
        UPDATE process_baselines SET last_healed_at = ?
        WHERE process_name = ?`, time.Now().UTC().Format(time.RFC3339), p.Name)

	h.db.Exec(`
    UPDATE process_baselines SET avg_memory = 0,avg_cpu = 0, sample_count = 0 WHERE process_name = ?`, p.Name)

	log.Printf("[healer] %s restarted with new pid %d", p.Name, newPID)
}
