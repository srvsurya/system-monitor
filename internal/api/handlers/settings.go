package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func GetUserSettings(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		var retention int
		var alertEmail *string
		err := db.Get(&alertEmail, `SELECT alert_email FROM users WHERE id = ?`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch email settings"})
			log.Printf("Failed to fetch user settings: %v", err)
			return
		}
		// retention settings
		err = db.Get(&retention, `SELECT retention_days FROM cleaner_settings WHERE id = 1`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
			log.Printf("Could not extract retention days for user:%v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"alert_email": alertEmail, "retention_days": retention})
	}
}

func UpdateUserSettings(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		log.Println("UserID:", userID)

		var body struct {
			AlertEmail string `json:"alert_email" binding:"required,email"`
			Retention  int    `json:"retention_days"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed at mapping json to struct"})
			log.Printf("Error at json bind:%v", err)
			return
		}

		_, err := db.Exec(`UPDATE users SET alert_email = ? WHERE id = ?`, body.AlertEmail, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
			log.Printf("Failed to update alert email: %v", err)
			return
		}
		_, err = db.Exec(`UPDATE cleaner_settings SET retention_days = ? WHERE id = 1`, body.Retention)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			log.Printf("Failed at retention updation:%v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"alert_email": body.AlertEmail, "retention_days": body.Retention})
	}
}
