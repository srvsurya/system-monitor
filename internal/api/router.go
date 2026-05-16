package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/srvsurya/system-monitor/internal/alerts"
	"github.com/srvsurya/system-monitor/internal/api/handlers"
	"github.com/srvsurya/system-monitor/internal/api/middleware"
)

func NewRouter(db *sqlx.DB, engine *alerts.Engine) *gin.Engine {
	r := gin.Default()
	hub := handlers.NewHub()
	go hub.Run()
	go hub.StartBroadcasting(db)
	// server health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	// route group for v1
	auth := r.Group("/auth")
	{
		auth.POST("/register", handlers.Register(db))
		auth.POST("/login", handlers.Login(db))
	}
	v1 := r.Group("api/v1")
	v1.Use(middleware.AuthRequired(db))
	{
		v1.GET("/me", handlers.Me(db))
		v1.POST("/logout", handlers.Logout(db))
		v1.GET("/stats", handlers.GetCurrentStats(db))
		v1.GET("/stats/history", handlers.GetStatsHistory(db))
		v1.GET("/ws", handlers.ServeWS(hub))
		v1.GET("/alerts", handlers.GetActiveAlerts(db))
		v1.GET("/alerts/history", handlers.GetAlertHistory(db))
		v1.POST("/alerts/rules", handlers.CreateRule(db, engine))
		v1.DELETE("/alerts/rules/:id", handlers.DeleteRule(db, engine))
	}
	return r
}
