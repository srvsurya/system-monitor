package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/srvsurya/system-monitor/internal/models"
)

// Function to get latest metrics - For live dashboard
func GetCurrentStats(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var metric models.SystemMetric
		err := db.Get(&metric, `SELECT * FROM system_metrics ORDER BY TIMESTAMP DESC LIMIT 1`)
		if err != nil {
			c.JSON(500, gin.H{"message": "Query Failed"})
			return
		}
		c.JSON(200, metric)
	}
}

// Function to get historical data - For insights and such
func GetStatsHistory(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// please remember to conv url params from str to int
		limitStr := c.DefaultQuery("limit", "50")
		fromStr := c.DefaultQuery("from", "")
		toStr := c.DefaultQuery("to", "")

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 50
		}
		// limit cap 1000 maybe lower?
		if limit > 1000 {
			limit = 1000
		}

		var metrics []models.SystemMetric

		if fromStr != "" && toStr != "" {
			from, err := time.Parse(time.RFC3339, fromStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' format, use RFC3339"})
				return
			}
			to, err := time.Parse(time.RFC3339, toStr) // RFC3339 - a tz format. look it up if you forget
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' format, use RFC3339"})
				return
			}

			err = db.Select(&metrics, `
				SELECT * FROM system_metrics
				WHERE timestamp BETWEEN $1 AND $2
				ORDER BY timestamp ASC
				LIMIT $3
			`, from, to, limit)
		} else {
			err = db.Select(&metrics, `
				SELECT * FROM system_metrics
				ORDER BY timestamp DESC
				LIMIT $1
			`, limit)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch history"})
			return
		}

		c.JSON(http.StatusOK, metrics)
	}
}
