package db

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func Connect() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", ""),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_NAME", ""),
	)
	var err error
	DB, err = sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection created successfully")
}
func RunMigrations() {
	m, err := migrate.New("file://internal/db/migrations",
		"postgres://monitor_user:user@localhost:5432/system_monitor?sslmode=disable")
	if err != nil {
		log.Fatalf("Migration initialization failed: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed:%v", err)
	}
	log.Println("Migrations applied succesfully")
}
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
