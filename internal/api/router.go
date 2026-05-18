package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/srvsurya/system-monitor/internal/alerts"
	"github.com/srvsurya/system-monitor/internal/api/handlers"
	"github.com/srvsurya/system-monitor/internal/api/middleware"
	"golang.org/x/time/rate"
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
		auth.POST("/register", middleware.RateLimit(rate.Every(time.Minute)/5, 5), handlers.Register(db))
		auth.POST("/login", middleware.RateLimit(rate.Every(time.Minute)/5, 5), handlers.Login(db))
	}
	v1 := r.Group("api/v1")
	v1.Use(middleware.AuthRequired(db))
	{
		v1.GET("/me", handlers.Me(db)) // test user details

		v1.POST("/logout", handlers.Logout(db))
		//metric routes
		v1.GET("/stats", middleware.RateLimit(rate.Every(time.Minute)/60, 5), handlers.GetCurrentStats(db))
		v1.GET("/stats/history", middleware.RateLimit(rate.Every(time.Minute)/60, 5), handlers.GetStatsHistory(db))
		// websocket init route
		v1.GET("/ws", handlers.ServeWS(hub))
		//alert routes
		v1.GET("/alerts", handlers.GetActiveAlerts(db))
		v1.GET("/alerts/history", handlers.GetAlertHistory(db))
		v1.POST("/alerts/rules", handlers.CreateRule(db, engine))
		v1.DELETE("/alerts/:id/rules", handlers.DeleteRule(db, engine))
		// process manager routes
		v1.POST("/processes/spawn", handlers.SpawnStressor(db)) // generates binary burner, for demo
		v1.GET("/processes", handlers.ListProcesses(db))
		v1.POST("/processes/:id/stop", middleware.RateLimit(rate.Every(time.Minute)/10, 5), handlers.StopProcess(db))
		v1.POST("/processes/:id/restart", middleware.RateLimit(rate.Every(time.Minute)/10, 5), handlers.RestartProcess(db))
	}
	return r
}
