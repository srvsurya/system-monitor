package cleaner

import (
	"log"

	"github.com/jmoiron/sqlx"
)

func RunRetentionPruner(db *sqlx.DB) {
	var retentionDays int
	err := db.QueryRow(`SELECT retention_days FROM cleaner_settings WHERE id = 1`).Scan(&retentionDays)
	if err != nil || retentionDays <= 0 {
		retentionDays = 5 // fallback default
	}

	tables := []struct {
		name  string
		query string
	}{ //  these are the tables to be wiped in the retention policy. Add or remove tables as you see fit or for future configuration
		{"system_metrics", `DELETE FROM system_metrics WHERE timestamp < datetime('now', '-' || ? || ' minutes')`},
		{"alerts", `DELETE FROM alerts WHERE status = 'resolved' AND triggered_at < datetime('now', '-' || ? || ' days')`},
		{"system_actions", `DELETE FROM system_actions WHERE created_at < datetime('now', '-' || ? || ' days')`},
	}

	for _, t := range tables {
		result, err := db.Exec(t.query, retentionDays)

		if err != nil {
			log.Printf("[pruner] failed to prune %s: %v", t.name, err)
			continue
		}
		log.Printf("Query:%v", result)
		rows, _ := result.RowsAffected()
		log.Printf("[pruner] pruned %d rows from %s", rows, t.name)
	}
}
