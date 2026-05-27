package models

import "time"

// for get endpoints
type SystemMetric struct {
	ID          int       `db:"id"           json:"id"`
	CPUUsage    float64   `db:"cpu_usage"    json:"cpu_usage"`
	MemoryUsed  int       `db:"memory_used"  json:"memory_used"`
	MemoryTotal int       `db:"memory_total" json:"memory_total"`
	DiskUsed    int       `db:"disk_used"    json:"disk_used"`
	DiskTotal   int       `db:"disk_total"   json:"disk_total"`
	NetUpload   float64   `db:"net_upload"   json:"net_upload"`
	NetDownload float64   `db:"net_download" json:"net_download"`
	Timestamp   time.Time `db:"timestamp"    json:"timestamp"`
}
type User struct {
	ID             int        `db:"id"              json:"id"`
	Name           string     `db:"name"            json:"name"`
	Email          string     `db:"email"           json:"email"`
	HashedPassword string     `db:"hashed_password" json:"-"`
	RegisteredOn   time.Time  `db:"registered_on"   json:"registered_on"`
	LastLogged     *time.Time `db:"last_logged"     json:"last_logged"`
	Verified       bool       `db:"verified" json:"verified"`
} // LastLogged can be NULL and it can't be represented with a regular time.Time, hence we use pointer

type Session struct {
	ID        int       `db:"id"         json:"id"`
	UserID    int       `db:"user_id"    json:"user_id"`
	Token     string    `db:"token"      json:"token"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
}

type AlertRule struct {
	ID              int       `db:"id"               json:"id"`
	Metric          string    `db:"metric"           json:"metric"`
	Operator        string    `db:"operator"         json:"operator"`
	Threshold       float64   `db:"threshold"        json:"threshold"`
	Enabled         bool      `db:"enabled"          json:"enabled"`
	DurationSeconds int       `db:"duration_seconds" json:"duration_seconds"`
	CreatedAt       time.Time `db:"created_at"       json:"created_at"`
}

type Alert struct {
	ID          int        `db:"id"           json:"id"`
	RuleID      int        `db:"rule_id"      json:"rule_id"`
	Value       float64    `db:"value"        json:"value"`
	Threshold   float64    `db:"threshold"    json:"threshold"`
	Status      bool       `db:"status"       json:"status"`
	TriggeredAt time.Time  `db:"triggered_at" json:"triggered_at"`
	ResolvedAt  *time.Time `db:"resolved_at"  json:"resolved_at"`
}

type ManagedProcess struct {
	ID        int       `db:"id"         json:"id"`
	PID       int       `db:"pid"        json:"pid"`
	Name      string    `db:"name"       json:"name"`
	Command   *string   `db:"command"    json:"command"`
	Status    string    `db:"status"     json:"status"`
	StartedAt time.Time `db:"started_at" json:"started_at"`
}

type SystemAction struct {
	ID          int       `db:"id"           json:"id"`
	ProcessID   *int      `db:"process_id"   json:"process_id"`
	ActionType  string    `db:"action_type"  json:"action_type"`
	Reason      string    `db:"reason"       json:"reason"`
	MetricValue float64   `db:"metric_value" json:"metric_value"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
}
