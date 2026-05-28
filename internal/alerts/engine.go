package alerts

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/srvsurya/system-monitor/internal/models"
)

// RuleState tracks in-memory evaluation state for a single alert rule.
type RuleState struct {
	Flagged          bool `json:"flagged"`
	ConsecutiveTicks int  `json:"consecutive_ticks"`
}

// Engine holds all alert rules and their current evaluation state.
type Engine struct {
	db      *sqlx.DB
	rules   []models.AlertRule
	state   map[int]RuleState
	onAlert func(rule models.AlertRule, value float64, emailAlert string)
}

// New creates a new Engine, loads rules from DB, and restores saved state if valid.
func New(db *sqlx.DB, onAlert func(models.AlertRule, float64, string)) *Engine {
	e := &Engine{
		db:      db,
		state:   make(map[int]RuleState),
		onAlert: onAlert,
	}
	e.loadRules()
	e.restoreState()
	return e
}

// loadRules fetches all enabled alert rules from the DB.
func (e *Engine) loadRules() {
	var rules []models.AlertRule
	err := e.db.Select(&rules, `SELECT * FROM alert_rules WHERE enabled = true`)
	if err != nil {
		log.Printf("[alerts] failed to load rules: %v", err)
		return
	}
	e.rules = rules
	log.Printf("[alerts] loaded %d rule(s)", len(rules))
}

// reload rules is for updating rules when the rules table is modified
func (e *Engine) ReloadRules() {
	e.loadRules()
}

// center piece of the engine btw
// Evaluate is called on every collector tick with the latest metric row.
// It checks each rule, updates state, fires alerts, and resolves them.
func (e *Engine) Evaluate(metric models.SystemMetric) {
	for _, rule := range e.rules {
		value := extractMetricValue(metric, rule.Metric)
		state := e.state[rule.ID]
		ticksRequired := ticksForDuration(rule.DurationSeconds)

		breached := evaluate(value, rule.Operator, rule.Threshold)

		// rundown: if on tick, breach is true - consecutive ticks go up and when met + flag is not triggered yet, the alert is set to fire.
		// once breach is false, alert is resolved - hence flag and ticks are reset.
		if breached {
			state.ConsecutiveTicks++
			if !state.Flagged && state.ConsecutiveTicks >= ticksRequired {
				// keep in mind: flag is for spam prevention mostly.
				state.Flagged = true
				e.fireAlert(rule, value)
			}
		} else {
			if state.Flagged {
				// Metric recovered — resolve the active alert
				e.resolveAlert(rule.ID)
			}
			// if breach is false, it doesn't matter if flag is true or false, just reset both regardless
			state.ConsecutiveTicks = 0
			state.Flagged = false
		}

		e.state[rule.ID] = state // feed state back to engine on every tick.
	}
}

// fireAlert writes an alert row to the DB and triggers the notification callback.
func (e *Engine) fireAlert(rule models.AlertRule, value float64) {
	_, err := e.db.Exec(`
		INSERT INTO alerts (rule_id, value, threshold, status)
		VALUES ($1, $2, $3, true)`,
		rule.ID, value, rule.Threshold,
	)
	if err != nil {
		log.Printf("[alerts] failed to insert alert for rule %d: %v", rule.ID, err)
		return
	}
	log.Printf("[alerts] FIRED: rule %d | %s %s %.2f (current: %.2f)",
		rule.ID, rule.Metric, rule.Operator, rule.Threshold, value)

	if e.onAlert != nil { // IF a callback exists somewhere, then only call it. in prod, just error handle or something
		if e.onAlert != nil {
			var alertEmail *string
			err := e.db.Get(&alertEmail, `
				SELECT alert_email FROM users 
				WHERE alert_email IS NOT NULL 
				LIMIT 1
			`)
			if err != nil || alertEmail == nil {
				log.Printf("[alerts] no alert email configured, skipping notification")
				return
			}
			e.onAlert(rule, value, *alertEmail)
		}
	}
}

// resolveAlert sets resolved_at and status=false on the most recent active alert for a rule.
func (e *Engine) resolveAlert(ruleID int) {
	_, err := e.db.Exec(`
		UPDATE alerts
		SET status = false, resolved_at = NOW()
		WHERE rule_id = $1 AND status = true`,
		ruleID,
	)
	if err != nil {
		log.Printf("[alerts] failed to resolve alert for rule %d: %v", ruleID, err)
		return
	}
	log.Printf("[alerts] RESOLVED: rule %d", ruleID)
}

// SaveStateToDB serialises current engine state and upserts it into alert_engine_state.
// Called on graceful shutdown.
func (e *Engine) SaveStateToDB() {
	data, err := json.Marshal(e.state)
	if err != nil {
		log.Printf("[alerts] failed to serialise state: %v", err)
		return
	}
	_, err = e.db.Exec(`
		INSERT INTO alert_engine_state (id, state_json, saved_at)
		VALUES (1, $1, NOW())
		ON CONFLICT (id) DO UPDATE
		SET state_json = EXCLUDED.state_json,
		    saved_at   = EXCLUDED.saved_at`,
		string(data), // Why? what's the use of having an extra conversion from json bytes to string before insertion? any reason, pls ask
	)
	if err != nil {
		log.Printf("[alerts] failed to save state: %v", err)
		return
	}
	log.Println("[alerts] state saved to DB")
}

// restoreState loads saved state from DB on startup.
// Discards it if saved_at is older than 24 hours.
func (e *Engine) restoreState() {
	var savedAt time.Time
	var stateJSON string

	err := e.db.QueryRow(`
		SELECT state_json, saved_at FROM alert_engine_state WHERE id = 1`,
	).Scan(&stateJSON, &savedAt)

	if err == sql.ErrNoRows {
		log.Println("[alerts] no saved state found, starting fresh")
		return
	}
	if err != nil {
		log.Printf("[alerts] failed to read saved state: %v", err)
		return
	}
	if time.Since(savedAt) > 24*time.Hour {
		log.Println("[alerts] saved state is stale (>24h), starting fresh")
		return
	}

	var restored map[int]RuleState
	if err := json.Unmarshal([]byte(stateJSON), &restored); err != nil {
		log.Printf("[alerts] failed to deserialise state: %v", err)
		return
	}

	e.state = restored
	log.Printf("[alerts] restored state for %d rule(s)", len(restored))
}

// extractMetricValue pulls the relevant field from a metric row by rule metric name.
func extractMetricValue(m models.SystemMetric, metric string) float64 {
	switch metric {
	case "cpu_usage":
		return m.CPUUsage
	case "memory_used":
		return float64(m.MemoryUsed)
	case "disk_used":
		return float64(m.DiskUsed)
	case "net_upload":
		return m.NetUpload
	case "net_download":
		return m.NetDownload
	default:
		return 0
	}
}

// evaluate applies the operator to value and threshold.
func evaluate(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	default:
		return false
	}
}

// ticksForDuration converts a duration in seconds to a tick count.
// Collector ticks every 5 seconds.
func ticksForDuration(seconds int) int {
	ticks := seconds / 5
	if ticks < 1 {
		return 1
	}
	return ticks
}
