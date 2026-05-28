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

		var alertEmail *string
		err := db.Get(&alertEmail, `SELECT alert_email FROM users WHERE id = $1`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
			log.Printf("Failed to fetch user settings: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"alert_email": alertEmail})
	}
}

func UpdateUserSettings(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		log.Println("UserID:", userID)

		var body struct {
			AlertEmail string `json:"alert_email" binding:"required,email"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Valid email required"})
			return
		}

		_, err := db.Exec(`UPDATE users SET alert_email = $1 WHERE id = $2`, body.AlertEmail, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
			log.Printf("Failed to update alert email: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"alert_email": body.AlertEmail})
	}
}
