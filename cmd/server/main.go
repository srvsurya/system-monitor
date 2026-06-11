package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"crypto/rand"
	"encoding/hex"

	"github.com/joho/godotenv"
	"github.com/srvsurya/system-monitor/internal/alerts"
	"github.com/srvsurya/system-monitor/internal/api"
	"github.com/srvsurya/system-monitor/internal/cleaner"
	"github.com/srvsurya/system-monitor/internal/collector"
	"github.com/srvsurya/system-monitor/internal/db"
	"github.com/srvsurya/system-monitor/internal/healer"
	"github.com/srvsurya/system-monitor/internal/logger"
	"github.com/srvsurya/system-monitor/internal/models"
	"github.com/srvsurya/system-monitor/internal/notify"
)

func generateRandomSecret() string { // For JWT in deployment
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func main() {
	err := godotenv.Load()
	// JWT secret passed down to router, then middleware and Login
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = generateRandomSecret()
	}
	//logger init
	zaplog, err := logger.New(true)
	if err != nil {
		panic(err)
	}
	defer zaplog.Sync()

	db.Connect()
	h := healer.New(db.DB)
	cleanerService := cleaner.New(db.DB)
	mailer := notify.New()
	alertEngine := alerts.New(db.DB, h, func(rule models.AlertRule, value float64, emailAlert string) {
		if err := mailer.SendAlert(emailAlert, rule.Metric, rule.Operator, rule.Threshold, value); err != nil {
			log.Println("Sending email failed:", err)
		}
	})
	cleaner.RunRetentionPruner(db.DB)
	r := api.NewRouter(db.DB, alertEngine, mailer, cleanerService, secret)
	// wrap gin inside http.Server so we can call the func Shutdown() on it
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	} // this is to handle server graceful shutdown

	col := collector.New(db.DB, alertEngine)
	// graceful shutdown of app goroutines using quit chan
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	col.Start()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()
	log.Println("The server is running....")

	<-quit
	// graceful app shutdown sequence
	log.Println("Shutting down....")
	alertEngine.SaveStateToDB() // save the alert engine state before shutdown so it can be restored on immediate startup
	log.Println("Saved state to database")
	col.Stop()

	// graceful server shutdown. why - because it waits until any pending or inflight requests finish processing before main returns. good practice, follow.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("Server has succesfully shutdown")

}
