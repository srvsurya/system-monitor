package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/srvsurya/system-monitor/internal/alerts"
	"github.com/srvsurya/system-monitor/internal/api"
	"github.com/srvsurya/system-monitor/internal/collector"
	"github.com/srvsurya/system-monitor/internal/db"
	"github.com/srvsurya/system-monitor/internal/logger"
	"github.com/srvsurya/system-monitor/internal/models"
	"github.com/srvsurya/system-monitor/internal/notify"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading env file")
	}
	//logger init
	zaplog, err := logger.New(true)
	if err != nil {
		panic(err)
	}
	defer zaplog.Sync()

	db.Connect()
	db.RunMigrations()
	mailer := notify.New()
	alertEngine := alerts.New(db.DB, func(rule models.AlertRule, value float64, emailAlert string) {
		if err := mailer.SendAlert(emailAlert, rule.Metric, rule.Operator, rule.Threshold, value); err != nil {
			log.Println("Sending email failed:", err)
		}
	})

	r := api.NewRouter(db.DB, alertEngine, mailer)
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
