package handlers

import (
	"net/http"
	"strconv"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/srvsurya/system-monitor/internal/alerts"
	"github.com/srvsurya/system-monitor/internal/models"
)

// GetActiveAlerts returns all alerts where status = true
func GetActiveAlerts(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var activeAlerts []models.Alert
		err := db.Select(&activeAlerts, `
			SELECT * FROM alerts WHERE status = true ORDER BY triggered_at DESC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch active alerts"})
			log.Printf("Failed to fetch active alerts: %v", err)
			return
		}
		c.JSON(http.StatusOK, activeAlerts)
	}
}

// GetAlertHistory returns all alerts regardless of status
func GetAlertHistory(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var allAlerts []models.Alert
		err := db.Select(&allAlerts, `
			SELECT * FROM alerts ORDER BY triggered_at DESC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch alert history"})
			log.Printf("Failed to fetch alert history: %v", err)
			return
		}
		c.JSON(http.StatusOK, allAlerts)
	}
}

// CreateRule inserts a new alert rule and reloads the engine
func CreateRule(db *sqlx.DB, engine *alerts.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Metric          string  `json:"metric"           binding:"required"`
			Operator        string  `json:"operator"         binding:"required"`
			Threshold       float64 `json:"threshold"        binding:"required"`
			DurationSeconds int     `json:"duration_seconds" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("Error at binding new rule response into json: %v", err)
			return
		}

		var rule models.AlertRule
		err := db.QueryRowx(`
			INSERT INTO alert_rules (metric, operator, threshold, duration_seconds)
			VALUES ($1, $2, $3, $4) ON CONFLICT (metric) DO UPDATE
			SET operator = EXCLUDED.operator,
			threshold = EXCLUDED.threshold,
			duration_seconds = EXCLUDED.duration_seconds
			RETURNING *`,
			input.Metric, input.Operator, input.Threshold, input.DurationSeconds,
		).StructScan(&rule)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create rule"})
			log.Printf("Failed to create rule:%v", err)
			return
		}

		engine.ReloadRules()
		c.JSON(http.StatusCreated, rule)
	}
}

// DeleteRule removes an alert rule and reloads the engine
func DeleteRule(db *sqlx.DB, engine *alerts.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
			log.Println("Invalid rule ID")
			return
		}
		db.Exec(`DELETE FROM alerts WHERE rule_id = $1`, id) // to jump over the foreign key constraint
		result, err := db.Exec(`DELETE FROM alert_rules WHERE id = $1`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete rule"})
			log.Printf("Failed to delete rule: %v", err)
			return
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
			log.Printf("Rule not found in DB: %v", err)
			return
		}

		engine.ReloadRules() // refer to engine.go, it's for pulling the latest rule table once the table has been modified.
		c.JSON(http.StatusOK, gin.H{"message": "rule deleted"})
	}
}

// Getting the list of rules for settings storage
func GetRules(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rules []models.AlertRule
		err := db.Select(&rules, `SELECT * FROM alert_rules ORDER BY id ASC`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch alert rules"})
			log.Printf("Failed to fetch alert rules: %v", err)
			return
		}
		c.JSON(http.StatusOK, rules)
	}
}
