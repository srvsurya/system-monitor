package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/srvsurya/system-monitor/internal/cleaner"
)

func Optimize(db *sqlx.DB, c *cleaner.Cleaner) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result := c.Optimize()
		log.Printf("[cleaner] optimize complete — killed %d, skipped %d", len(result.Killed), result.Skipped)
		ctx.JSON(http.StatusOK, result)
	}
}

func AddToIgnoreList(db *sqlx.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body struct {
			ProcessName string `json:"process_name" binding:"required"`
		}
		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "process_name required"})
			return
		}
		_, err := db.Exec(`
            INSERT INTO cleaner_ignore_list (process_name)
            VALUES (?)
            ON CONFLICT(process_name) DO NOTHING`, body.ProcessName)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add to ignore list"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "added to ignore list"})
	}
}

func GetIgnoreList(db *sqlx.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var list []struct {
			ID          int    `db:"id"           json:"id"`
			ProcessName string `db:"process_name" json:"process_name"`
		}
		err := db.Select(&list, `SELECT id, process_name FROM cleaner_ignore_list ORDER BY added_at DESC`)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch ignore list"})
			return
		}
		ctx.JSON(http.StatusOK, list)
	}
}

func RemoveFromIgnoreList(db *sqlx.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		db.Exec(`DELETE FROM cleaner_ignore_list WHERE id = ?`, id)
		ctx.JSON(http.StatusOK, gin.H{"message": "removed from ignore list"})
	}
}
