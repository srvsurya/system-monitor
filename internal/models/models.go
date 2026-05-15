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
} // LastLogged can be NULL and it can't be represented with a regular time.Time, hence we use pointer

type Session struct {
	ID        int       `db:"id"         json:"id"`
	UserID    int       `db:"user_id"    json:"user_id"`
	Token     string    `db:"token"      json:"token"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
}
