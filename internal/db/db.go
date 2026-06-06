package db

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var DB *sqlx.DB

func Connect() {
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "monitor.db"
	}
	var err error
	DB, err = sqlx.Connect("sqlite", path)
	if err != nil {
		log.Fatalf("Failed to connect to SQLite: %v", err)
	}
	DB.SetMaxOpenConns(1)
	DB.Exec("PRAGMA journal_mode=WAL;")  // for less aggressive locking by default, choose write head logging for better concurrency
	DB.Exec("PRAGMA busy_timeout=5000;") // fallback
	DB.Exec("PRAGMA foreign_keys=ON;")   // make sure foreign key constraint violations are enabled
	initSchema()
	DB.Exec(`UPDATE alerts SET status = 0 WHERE status = 1`) // cleaning stale alerts
	log.Println("Cleaned up stale alerts")
	log.Println("SQLite connected and schema ready")
}

func initSchema() { // migrations moved to initSchema
	schema := `
    CREATE TABLE IF NOT EXISTS users (
        id               INTEGER PRIMARY KEY AUTOINCREMENT,
        name             TEXT,
        email            TEXT UNIQUE,
        hashed_password  TEXT,
        verified         INTEGER NOT NULL DEFAULT 0,
        registered_on    TIMESTAMP DEFAULT (datetime('now')),
        last_logged      TIMESTAMP,
        alert_email      TEXT
    );
    CREATE TABLE IF NOT EXISTS sessions (
        id          INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id     INTEGER REFERENCES users(id) ON DELETE CASCADE,
        token       TEXT UNIQUE,
        created_at  TIMESTAMP DEFAULT (datetime('now')),
        expires_at  TIMESTAMP
    );
    CREATE TABLE IF NOT EXISTS system_metrics (
        id            INTEGER PRIMARY KEY AUTOINCREMENT,
        cpu_usage     REAL,
        memory_used   INTEGER,
        memory_total  INTEGER,
        disk_used     INTEGER,
        disk_total    INTEGER,
        net_upload    REAL,
        net_download  REAL,
        timestamp     TIMESTAMP DEFAULT (datetime('now'))
    );
    CREATE TABLE IF NOT EXISTS alert_rules (
        id               INTEGER PRIMARY KEY AUTOINCREMENT,
        metric           TEXT UNIQUE,
        operator         TEXT,
        threshold        REAL,
        enabled          INTEGER DEFAULT 1,
        duration_seconds INTEGER,
        created_at       TIMESTAMP DEFAULT (datetime('now'))
    );
    CREATE TABLE IF NOT EXISTS alerts (
        id           INTEGER PRIMARY KEY AUTOINCREMENT,
        rule_id      INTEGER REFERENCES alert_rules(id),
        value        REAL,
        threshold    REAL,
        status       INTEGER DEFAULT 1,
        triggered_at TIMESTAMP DEFAULT (datetime('now')),
        resolved_at  TIMESTAMP
    );
    CREATE TABLE IF NOT EXISTS managed_processes (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        pid        INTEGER,
        name       TEXT,
        command    TEXT,
        status     TEXT,
        pinned     INTEGER DEFAULT 0,
        started_at TIMESTAMP DEFAULT (datetime('now'))
    );
    CREATE TABLE IF NOT EXISTS system_actions (
        id           INTEGER PRIMARY KEY AUTOINCREMENT,
        process_id   INTEGER REFERENCES managed_processes(id) ON DELETE CASCADE,
        action_type  TEXT,
        reason       TEXT,
        metric_value REAL,
        created_at   TIMESTAMP DEFAULT (datetime('now'))
    );
    CREATE TABLE IF NOT EXISTS alert_engine_state (
        id         INTEGER PRIMARY KEY DEFAULT 1,
        state_json TEXT NOT NULL,
        saved_at   TIMESTAMP NOT NULL DEFAULT (datetime('now'))
    );
    CREATE TABLE IF NOT EXISTS verification_tokens (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id    INTEGER REFERENCES users(id) ON DELETE CASCADE,
        token      TEXT UNIQUE NOT NULL,
        created_at TIMESTAMP DEFAULT (datetime('now')),
        expires_at TIMESTAMP NOT NULL
    );
        CREATE TABLE IF NOT EXISTS process_baselines (
        id             INTEGER PRIMARY KEY AUTOINCREMENT,
        process_name   TEXT UNIQUE,
        sample_count   INTEGER DEFAULT 0,
        avg_cpu        REAL DEFAULT 0,
        avg_memory     REAL DEFAULT 0,
        last_healed_at TIMESTAMP,
        last_updated   TIMESTAMP DEFAULT (datetime('now'))
    );
        CREATE TABLE IF NOT EXISTS cleaner_ignore_list (
        id           INTEGER PRIMARY KEY AUTOINCREMENT,
        process_name TEXT UNIQUE,
        added_at     TIMESTAMP DEFAULT (datetime('now'))
    );
        CREATE TABLE IF NOT EXISTS cleaner_settings (
        id                  INTEGER PRIMARY KEY DEFAULT 1,
        cpu_threshold       REAL DEFAULT 80.0,
        mem_threshold_mb    REAL DEFAULT 500.0,
        duplicate_threshold INTEGER DEFAULT 3,
        smart_heal_enabled  INTEGER DEFAULT 0,
        updated_at          TIMESTAMP DEFAULT (datetime('now'))
    );
        INSERT OR IGNORE INTO cleaner_settings (id) VALUES (1);`

	DB.MustExec(schema)

}
