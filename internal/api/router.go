package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/srvsurya/system-monitor/internal/alerts"
	"github.com/srvsurya/system-monitor/internal/api/handlers"
	"github.com/srvsurya/system-monitor/internal/api/middleware"
	"github.com/srvsurya/system-monitor/internal/notify"
	"golang.org/x/time/rate"
)

func NewRouter(db *sqlx.DB, engine *alerts.Engine, mailer *notify.Mailer) *gin.Engine {
	r := gin.Default()
	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
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
		auth.POST("/register", middleware.RateLimit(rate.Every(time.Minute)/10, 5), handlers.Register(db, mailer))
		auth.POST("/login", middleware.RateLimit(rate.Every(time.Minute)/10, 5), handlers.Login(db))
		auth.GET("/verify", middleware.RateLimit(rate.Every(time.Minute)/10, 5), handlers.VerifyEmail(db))
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
		v1.GET("/alerts/rules", handlers.GetRules(db))
		v1.DELETE("/alerts/rules/:id", handlers.DeleteRule(db, engine))
		// process manager routes
		v1.POST("/processes/spawn", handlers.SpawnStressor(db)) // generates binary burner, for demo
		v1.GET("/processes", handlers.ListProcesses(db))
		v1.POST("/processes/stop/:id", middleware.RateLimit(rate.Every(time.Minute)/10, 5), handlers.StopProcess(db))
		v1.POST("/processes/restart/:id", middleware.RateLimit(rate.Every(time.Minute)/10, 5), handlers.RestartProcess(db))
		v1.POST("/processes/register", handlers.RegisterProcess(db))
	}
	return r
}
