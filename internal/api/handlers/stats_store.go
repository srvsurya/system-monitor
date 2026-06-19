package handlers

import (
	"github.com/jmoiron/sqlx"
	"github.com/srvsurya/system-monitor/internal/models"
)

type StatsStore interface {
	GetLatestMetric() (models.SystemMetric, error)
}

type SQLStatsStore struct {
	DB *sqlx.DB
}

func (s SQLStatsStore) GetLatestMetric() (models.SystemMetric, error) {
	var metric models.SystemMetric

	err := s.DB.Get(
		&metric,
		`SELECT * FROM system_metrics ORDER BY timestamp DESC LIMIT 1`,
	)

	return metric, err
}
