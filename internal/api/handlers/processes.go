package handlers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"

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
			return
		}
		var managed models.ManagedProcess
		err = db.Get(&managed, `SELECT * FROM managed_processes WHERE id = $1`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Process not in the managed process list"})
			return
		}
		pid := managed.PID
		process, err := os.FindProcess(pid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find process"})
			return
		}
		if err := process.Kill(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop process"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Process stopped"})
		db.Exec(`UPDATE managed_processes SET status = stopped WHERE id = $1`, id)                                                    // Update the table in db after stopping the process
		db.Exec(`INSERT INTO system_actions(process_id,action_type,reason) VALUES($1,$2,$3)`, pid, "Stop", "Process stopped via API") // Log into the system_actions table

	}
}
func RestartProcess(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		// sandbox check
		var managed models.ManagedProcess
		err = db.Get(&managed, `SELECT * FROM managed_processes WHERE id = $1`, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "process not in managed list"})
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
			return
		}

		c.JSON(http.StatusCreated, managed)
	}
}
