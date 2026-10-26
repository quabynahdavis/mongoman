package sqlite

// SchemaSQL contains the SQLite DDL for mongoman's database.
const SchemaSQL = `
-- Instances table: one row per MongoDB instance
CREATE TABLE IF NOT EXISTS instances (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL UNIQUE,
    port        INTEGER NOT NULL,
    db_path     TEXT NOT NULL,
    log_path    TEXT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    updated_at  TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    deleted_at  TIMESTAMP
);

-- Launch history: one row per launch session
CREATE TABLE IF NOT EXISTS launch_history (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    instance_id INTEGER NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    start_time  TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    end_time    TIMESTAMP,
    duration    INTEGER, -- seconds, computed when end_time is set
    created_at  TIMESTAMP NOT NULL DEFAULT (datetime('now'))
);

-- Config table: key-value store for global settings
CREATE TABLE IF NOT EXISTS config (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_instances_name ON instances(name);
CREATE INDEX IF NOT EXISTS idx_launch_history_instance ON launch_history(instance_id);
CREATE INDEX IF NOT EXISTS idx_launch_history_time ON launch_history(start_time);

-- Trigger to update instances.updated_at on changes
CREATE TRIGGER IF NOT EXISTS trg_instances_updated_at
AFTER UPDATE ON instances
BEGIN
    UPDATE instances SET updated_at = datetime('now') WHERE id = NEW.id;
END;
`