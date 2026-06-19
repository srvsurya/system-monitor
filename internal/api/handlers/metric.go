package handlers

import (
	"net/http"
	"strconv"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/srvsurya/system-monitor/internal/models"
)

// Function to get latest metrics - For live dashboard
func GetCurrentStats(store StatsStore) gin.HandlerFunc {
	return func(c *gin.Context) {

		metric, err := store.GetLatestMetric()

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
			from, err := time.Parse("2006-01-02T15:04:05", fromStr)
			if err != nil {
				from, err = time.Parse(time.RFC3339, fromStr)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' format"})
					log.Printf("Invalid from query format:%v", err)
					return
				}
			}
			to, err := time.Parse("2006-01-02T15:04:05", toStr)
			if err != nil {
				to, err = time.Parse(time.RFC3339, toStr)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' format"})
					log.Printf("Invalid to query format:%v", err)
					return
				}
			}
			fromFmt := from.UTC().Format("2006-01-02 15:04:05")
			toFmt := to.UTC().Format("2006-01-02 15:04:05")

			err = db.Select(&metrics, `
				SELECT * FROM system_metrics
				WHERE timestamp BETWEEN ? AND ?
				ORDER BY timestamp ASC
				LIMIT ?
			`, fromFmt, toFmt, limit)
		} else {
			err = db.Select(&metrics, `
				SELECT * FROM system_metrics
				ORDER BY timestamp DESC
				LIMIT ?
			`, limit)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch history"})
			log.Printf("Failed to fetch history:%v", err)
			return
		}

		c.JSON(http.StatusOK, metrics)
	}
}
