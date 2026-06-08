package cleaner

import (
	"log"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/shirou/gopsutil/v3/process"
)

type CleanerSettings struct {
	CPUThreshold       float64 `db:"cpu_threshold"`
	MemThresholdMB     float64 `db:"mem_threshold_mb"`
	DuplicateThreshold int     `db:"duplicate_threshold"`
}

func (c *Cleaner) loadSettings() CleanerSettings {
	var s CleanerSettings
	err := c.db.Get(&s, `SELECT cpu_threshold, mem_threshold_mb, duplicate_threshold FROM cleaner_settings WHERE id = 1`)
	if err != nil {
		// fallback to safe defaults if table read fails
		return CleanerSettings{
			CPUThreshold:       80.0,
			MemThresholdMB:     500.0,
			DuplicateThreshold: 3,
		}
	}
	return s
}

// processes that should never be touched
var protectedNames = map[string]bool{
	"systemd": true, "init": true, "kthreadd": true,
	"kworker": true, "ksoftirqd": true, "rcu_sched": true,
	"bash": true, "sh": true, "ssh": true, "sshd": true,
	"sudo": true, "su": true, "login": true,
}

type KilledProcess struct {
	PID    int32   `json:"pid"`
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory_mb"`
	Reason string  `json:"reason"`
}

type OptimizeResult struct {
	Killed  []KilledProcess `json:"killed"`
	Skipped int             `json:"skipped"`
	Errors  []string        `json:"errors"`
}

type Cleaner struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Cleaner {
	return &Cleaner{db: db}
}

func (c *Cleaner) Optimize() OptimizeResult {
	result := OptimizeResult{
		Killed: []KilledProcess{},
		Errors: []string{},
	}
	settings := c.loadSettings()
	// load ignore list from DB
	var ignoreList []string
	c.db.Select(&ignoreList, `SELECT process_name FROM cleaner_ignore_list`)

	procs, err := process.Processes()
	if err != nil {
		result.Errors = append(result.Errors, "failed to fetch processes")
		return result
	}

	// track process name counts for duplicate detection
	nameCounts := map[string]int{}
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		nameCounts[name]++
	}

	for _, p := range procs {
		name, _ := p.Name()

		// skip protected
		if isProtected(p, name, ignoreList) {
			result.Skipped++
			continue
		}

		// check zombie
		statuses, _ := p.Status()
		isZombie := len(statuses) > 0 && statuses[0] == "zombie"

		// check cpu/mem
		cpu, _ := p.CPUPercent()
		memInfo, _ := p.MemoryInfo()
		memMB := float64(0)
		if memInfo != nil {
			memMB = float64(memInfo.RSS) / 1024 / 1024
		}

		cpuHog := cpu > settings.CPUThreshold
		memHog := memMB > settings.MemThresholdMB
		isDuplicate := nameCounts[name] > 3 && !isSystemProcess(name)

		reason := ""
		if isZombie {
			reason = "zombie process"
		} else if cpuHog && memHog {
			reason = "cpu and memory hog"
		} else if cpuHog {
			reason = "cpu hog"
		} else if memHog {
			reason = "memory hog"
		} else if isDuplicate {
			reason = "duplicate process"
		}

		if reason == "" {
			result.Skipped++
			continue
		}

		// kill it
		proc, err := os.FindProcess(int(p.Pid))
		if err != nil {
			result.Errors = append(result.Errors, "could not find process "+name)
			continue
		}
		if err := proc.Kill(); err != nil {
			result.Errors = append(result.Errors, "failed to kill "+name)
			continue
		}

		log.Printf("[cleaner] killed %s (pid %d) — %s", name, p.Pid, reason)
		result.Killed = append(result.Killed, KilledProcess{
			PID:    p.Pid,
			Name:   name,
			CPU:    cpu,
			Memory: memMB,
			Reason: reason,
		})
	}

	return result
}

func isProtected(p *process.Process, name string, ignoreList []string) bool {
	// low PID = kernel/system process
	if p.Pid < 500 {
		return true
	}
	// protected names
	if protectedNames[strings.ToLower(name)] {
		return true
	}
	// user ignore list
	for _, ignored := range ignoreList {
		if strings.EqualFold(name, ignored) {
			return true
		}
	}
	return false
}

func isSystemProcess(name string) bool {
	systemPrefixes := []string{"kworker", "kthread", "rcu", "migration", "watchdog", "gnome", "gvfs", "ibus", "dbus", "xdg", "gdm", "systemd"}
	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(strings.ToLower(name), prefix) {
			return true
		}
	}
	return false
}
