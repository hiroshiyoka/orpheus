CREATE TABLE IF NOT EXISTS projects (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    name               TEXT NOT NULL,
    url                TEXT NOT NULL,
    cloudflare_zone_id TEXT,
    check_interval_seconds INTEGER DEFAULT 60,
    is_active          BOOLEAN DEFAULT 1,
    created_at         DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS checks (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id       INTEGER NOT NULL REFERENCES projects(id),
    checked_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_up            BOOLEAN NOT NULL,
    status_code      INTEGER,
    response_time_ms INTEGER,
    error_message    TEXT
);
CREATE INDEX IF NOT EXISTS idx_checks_project_time ON checks(project_id, checked_at);

CREATE TABLE IF NOT EXISTS metrics (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id      INTEGER NOT NULL REFERENCES projects(id),
    period_start    DATETIME NOT NULL,
    requests_count  INTEGER DEFAULT 0,
    error_count     INTEGER DEFAULT 0,
    p50_response_ms INTEGER,
    p99_response_ms INTEGER
);
CREATE INDEX IF NOT EXISTS idx_metrics_project_time ON metrics(project_id, period_start);

CREATE TABLE IF NOT EXISTS incidents (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id   INTEGER NOT NULL REFERENCES projects(id),
    type         TEXT NOT NULL,
    started_at   DATETIME NOT NULL,
    resolved_at  DATETIME,
    description  TEXT
);
CREATE INDEX IF NOT EXISTS idx_incidents_project ON incidents(project_id, started_at);

CREATE TABLE IF NOT EXISTS alerts_sent (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    incident_id INTEGER REFERENCES incidents(id),
    channel     TEXT DEFAULT 'telegram',
    message     TEXT,
    sent_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);
