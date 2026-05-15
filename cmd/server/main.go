package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/srvsurya/system-monitor/internal/api"
	"github.com/srvsurya/system-monitor/internal/collector"
	"github.com/srvsurya/system-monitor/internal/db"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading env file")
	}
	fmt.Println("System monitor is starting....")
	db.Connect()
	db.RunMigrations()

	r := api.NewRouter(db.DB)

	col := collector.New(db.DB)
	col.Start()
	defer col.Stop()
	r.Run(":8080")
}
