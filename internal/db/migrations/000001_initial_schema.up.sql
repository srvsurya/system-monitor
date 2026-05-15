-- system_metrics
CREATE TABLE IF NOT EXISTS system_metrics (
    id           SERIAL PRIMARY KEY,
    cpu_usage    FLOAT,
    memory_used  INT,
    memory_total INT,
    disk_used    INT,
    disk_total   INT,
    net_upload   FLOAT,
    net_download FLOAT,
    timestamp    TIMESTAMP DEFAULT NOW()
);

-- alert_rules
CREATE TABLE IF NOT EXISTS alert_rules (
    id               SERIAL PRIMARY KEY,
    metric           VARCHAR(50),
    operator         VARCHAR(10),
    threshold        FLOAT,
    enabled          BOOLEAN DEFAULT TRUE,
    duration_seconds INT,
    created_at       TIMESTAMP DEFAULT NOW()
);

-- alerts (metric removed, status changed to BOOLEAN)
CREATE TABLE IF NOT EXISTS alerts (
    id           SERIAL PRIMARY KEY,
    rule_id      INT REFERENCES alert_rules(id),
    value        FLOAT,
    threshold    FLOAT,
    status       BOOLEAN DEFAULT TRUE,
    triggered_at TIMESTAMP DEFAULT NOW(),
    resolved_at  TIMESTAMP
);

-- managed_processes
CREATE TABLE IF NOT EXISTS managed_processes (
    id         SERIAL PRIMARY KEY,
    pid        INT,
    name       VARCHAR(100),
    status     VARCHAR(20),
    started_at TIMESTAMP DEFAULT NOW()
);

-- users
CREATE TABLE IF NOT EXISTS users (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(100),
    email           VARCHAR(150) UNIQUE,
    hashed_password TEXT,
    registered_on   TIMESTAMP DEFAULT NOW(),
    last_logged     TIMESTAMP
);

-- smart_heal_config
CREATE TABLE IF NOT EXISTS smart_heal_config (
    id                      SERIAL PRIMARY KEY,
    user_id                 INT REFERENCES users(id),
    enabled                 BOOLEAN DEFAULT TRUE,
    cpu_threshold           FLOAT,
    cpu_duration_seconds    INT,
    memory_threshold        FLOAT,
    memory_duration_seconds INT,
    updated_at              TIMESTAMP DEFAULT NOW()
);

-- system_actions
CREATE TABLE IF NOT EXISTS system_actions (
    id           SERIAL PRIMARY KEY,
    process_id   INT REFERENCES managed_processes(id),
    action_type  VARCHAR(50),
    reason       VARCHAR(100),
    metric_value FLOAT,
    created_at   TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_metrics_timestamp
    ON system_metrics(timestamp);

CREATE INDEX IF NOT EXISTS idx_metrics_timestamp_metric
    ON system_metrics(timestamp);

CREATE INDEX IF NOT EXISTS idx_alerts_status
    ON alerts(status);