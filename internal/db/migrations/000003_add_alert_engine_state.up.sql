CREATE TABLE IF NOT EXISTS alert_engine_state (
    id           INTEGER PRIMARY KEY DEFAULT 1,
    state_json   TEXT NOT NULL,
    saved_at     TIMESTAMP NOT NULL DEFAULT NOW()
);