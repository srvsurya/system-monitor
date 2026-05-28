package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/srvsurya/system-monitor/internal/models"
)

func ListProcesses(db *sqlx.DB) gin.HandlerFunc { // list out processes (all)
	return func(c *gin.Context) {
		proc, err := process.Processes()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to fetch the list of processes"})
			log.Printf("Failed to fetch process list:%v", err)
			return
		}
		type ProcessInfo struct {
			PID    int32   `json:"pid"`
			Name   string  `json:"name"`
			CPU    float64 `json:"cpu_percentage"`
			Memory float32 `json:"memory_percentage"`
		}
		var list []ProcessInfo

		for _, p := range proc {
			name, err := p.Name()
			if err != nil {
				continue
			}
			cpu, err := p.CPUPercent()
			memory, err := p.MemoryPercent()
			list = append(list, ProcessInfo{
				PID:    p.Pid,
				Name:   name,
				CPU:    cpu,
				Memory: memory,
			})
		}
		c.JSON(http.StatusOK, list)

	}

}
func StopProcess(db *sqlx.DB) gin.HandlerFunc { // stop process ONLY from managed
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid ID"})
			log.Printf("Invalid ID from query params:%v", err)
			return
		}
		var managed models.ManagedProcess
		err = db.Get(&managed, `SELECT * FROM managed_processes WHERE id = $1`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Process not in the managed process list"})
			log.Printf("Process not in managed list")
			return
		}
		pid := managed.PID
		process, err := os.FindProcess(pid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find process"})
			log.Printf("Process does not exist in the OS")
			return
		}
		if err := process.Kill(); err != nil { // check graceful shutdown viability
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop process"})
			log.Printf("Failed to stop process: %v", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Process stopped"})
		db.Exec(`UPDATE managed_processes SET status = 'stopped' WHERE id = $1`, id)                                                  // Update the table in db after stopping the process
		db.Exec(`INSERT INTO system_actions(process_id,action_type,reason) VALUES($1,$2,$3)`, pid, "Stop", "Process stopped via API") // Log into the system_actions table

	}
}
func RestartProcess(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			log.Printf("Invalid ID")
			return
		}

		// sandbox check
		var managed models.ManagedProcess
		err = db.Get(&managed, `SELECT * FROM managed_processes WHERE id = $1`, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "process not in managed list"})
			log.Printf("Process not managed in the list")
			return
		}

		// kill existing
		proc, err := os.FindProcess(managed.PID)
		if err == nil {
			proc.Kill()
		}

		// Personal Note: After process kill, the PID resets. We restart using the registered process command stored in the db, this will stay persistent.
		// Personal Note: For now, ONLY stressor can be in managed processes. Custom managed processes scope in the later weeks.
		cmd := exec.Command("./stressor") // executable CPU burner
		if err := cmd.Start(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restart process"})
			log.Printf("Failed to restart process: %v", err)
			return
		}

		newPID := cmd.Process.Pid

		// update managed_processes with new PID and status
		db.Exec(`
			UPDATE managed_processes
			SET pid = $1, status = 'running'
			WHERE id = $2`, newPID, id)

		// log the action
		db.Exec(`
			INSERT INTO system_actions (process_id, action_type, reason)
			VALUES ($1, 'restart', 'manual restart via API')`, id)

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("process restarted with new pid %d", newPID),
			"pid":     newPID,
		})
	}
}

// SpawnStressor starts the stressor binary and registers it in managed_processes
func SpawnStressor(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		cmd := exec.Command("./stressor")
		if err := cmd.Start(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to spawn stressor"})
			log.Printf("failed to create the stressor script: %v", err)
			return
		}

		pid := cmd.Process.Pid

		var managed models.ManagedProcess
		err := db.QueryRowx(`
			INSERT INTO managed_processes (pid, name, status)
			VALUES ($1, 'stressor', 'running')
			RETURNING *`, pid,
		).StructScan(&managed)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register process"})
			log.Printf("Failed to register process: %v", err)
			return
		}

		c.JSON(http.StatusCreated, managed)
	}
}
func RegisterProcess(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		pid, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Server error"})
			log.Printf("String conversion error: %v", err)
			return
		}
		proc, err := process.NewProcess(int32(pid))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Internal Server Error"})
			log.Printf("error:%v", err)
			return
		}
		name, err := proc.Name()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			log.Printf("error:%v", err)
			return
		}
		cmdline, err := proc.Cmdline()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			log.Printf("error:%v", err)
			return
		}
		p := models.ManagedProcess{
			PID:       int(proc.Pid),
			Name:      name,
			Command:   &cmdline,
			Status:    "running",
			StartedAt: time.Now(),
		}

		req := models.ManagedProcess{}

		err = db.Get(&req, `SELECT * FROM managed_processes WHERE name = $1`, name)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Process is already being managed"})
			log.Printf("Process already in the managed_processes table")
			return
		} else if err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			log.Printf("error:%v", err)
			return
		}
		var id int
		// db insert
		err = db.QueryRow(`INSERT INTO managed_processes(pid,name,command,status,started_at) VALUES($1,$2,$3,$4,$5) RETURNING id`, p.PID, p.Name, p.Command, p.Status, p.StartedAt).Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Failure"})
			log.Printf("DB error:%v", err)
			return
		}
		p.ID = id
		c.JSON(http.StatusCreated, p)

	}
}
func GetManagedProcesses(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var processes []models.ManagedProcess
		type managed struct {
			models.ManagedProcess
			CPU    float64 `json:"cpu_percentage"`
			Memory float32 `json:"memory_percentage"`
		}
		var info []managed

		err := db.Select(&processes, `SELECT * FROM managed_processes`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal DB error"})
			log.Printf("DB error:%v", err)
			return
		}
		for _, p := range processes {
			cpu := 0.0
			mem := float32(0.0)

			if p.Status == "running" {
				proc, err := process.NewProcess(int32(p.PID))
				if err == nil {
					cpu, _ = proc.CPUPercent()
					mem, _ = proc.MemoryPercent()
				}
			}

			info = append(info, managed{
				ManagedProcess: p,
				CPU:            cpu,
				Memory:         mem,
			})
		}
		c.JSON(http.StatusOK, info)

	}
}
func UpdatePinnedStatus(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			log.Printf("Conversion error:%v", err)
			return
		}
		db.Exec(`UPDATE managed_processes SET pinned = NOT pinned WHERE pid = $1`, id)

		c.JSON(http.StatusOK, gin.H{"message": "Pinned status updated"})
	}
}
