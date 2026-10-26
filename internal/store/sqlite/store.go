package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/quabynahdavis/mongoman/internal/config"
)

// Store wraps the SQLite database and provides CRUD for instances.
type Store struct {
	db    *sql.DB
	paths *config.Paths
}

// NewStore opens (or creates) the SQLite database at the standard location.
// The database file lives in the config directory.
func NewStore(paths *config.Paths) (*Store, error) {
	dbPath := filepath.Join(paths.ConfigDir, "mongoman.db")

	// Ensure the config directory exists
	if err := os.MkdirAll(paths.ConfigDir, 0o755); err != nil {
		return nil, fmt.Errorf("cannot create config dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?mode=rwc")
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("cannot enable foreign keys: %w", err)
	}

	// Apply schema
	if _, err := db.Exec(SchemaSQL); err != nil {
		return nil, fmt.Errorf("cannot apply schema: %w", err)
	}

	s := &Store{db: db, paths: paths}

	// Run migration from old JSON files
	if err := s.migrateFromJSON(); err != nil {
		return nil, fmt.Errorf("JSON migration failed: %w", err)
	}

	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// migrateFromJSON scans the old JSON metadata directory and imports any
// instances not already present in the database.
func (s *Store) migrateFromJSON() error {
	entries, err := os.ReadDir(s.paths.ConfigDir)
	if err != nil {
		return fmt.Errorf("cannot read config dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || len(e.Name()) < 6 || e.Name()[len(e.Name())-5:] != ".json" {
			continue
		}
		name := e.Name()[:len(e.Name())-5]

		// Skip if already in DB
		if _, err := s.GetInstanceByName(name); err == nil {
			continue
		}

		// Read the JSON file
		jsonPath := filepath.Join(s.paths.ConfigDir, e.Name())
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			// Skip unreadable files
			continue
		}

		var meta struct {
			Name          string    `json:"name"`
			Port          int       `json:"port"`
			CreatedAt     time.Time `json:"created_at"`
			LaunchCount   int       `json:"launch_count"`
			LaunchHistory []struct {
				Start time.Time  `json:"start"`
				End   *time.Time `json:"end,omitempty"`
			} `json:"launch_history"`
			DeletedAt *time.Time `json:"deleted_at,omitempty"`
		}
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}

		// Create instance in DB
		dbPath := s.paths.DBPath(name)
		logPath := s.paths.LogPath(name)
		id, err := s.CreateInstance(name, meta.Port, dbPath, logPath)
		if err != nil {
			continue
		}

		// Update created_at if different
		if !meta.CreatedAt.IsZero() {
			s.db.Exec(`UPDATE instances SET created_at = ? WHERE id = ?`, meta.CreatedAt, id)
		}

		// Migrate launch history
		for _, lr := range meta.LaunchHistory {
			if lr.Start.IsZero() {
				continue
			}
			recID, err := s.RecordLaunch(id)
			if err != nil {
				continue
			}
			if lr.End != nil {
				s.db.Exec(
					`UPDATE launch_history SET end_time = ?, duration = CAST(strftime('%s', ?) - strftime('%s', start_time) AS INTEGER) WHERE id = ?`,
					*lr.End, *lr.End, recID,
				)
			}
		}

		// Set deleted_at if present
		if meta.DeletedAt != nil {
			s.db.Exec(`UPDATE instances SET deleted_at = ? WHERE id = ?`, *meta.DeletedAt, id)
		}
	}

	return nil
}

// Instance represents a MongoDB instance in the database.
type Instance struct {
	ID        int64
	Name      string
	Port      int
	DBPath    string
	LogPath   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// LaunchRecord represents a single launch session.
type LaunchRecord struct {
	ID          int64
	InstanceID  int64
	StartTime   time.Time
	EndTime     *time.Time
	DurationSec *int64 // seconds, computed when EndTime is set
}

// CreateInstance inserts a new instance and returns its ID.
func (s *Store) CreateInstance(name string, port int, dbPath, logPath string) (int64, error) {
	now := time.Now().UTC()
	result, err := s.db.Exec(
		`INSERT INTO instances (name, port, db_path, log_path, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		name, port, dbPath, logPath, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("cannot create instance: %w", err)
	}
	id, _ := result.LastInsertId()
	return id, nil
}

// GetInstanceByName retrieves an instance by name.
func (s *Store) GetInstanceByName(name string) (*Instance, error) {
	row := s.db.QueryRow(
		`SELECT id, name, port, db_path, log_path, created_at, updated_at, deleted_at
		 FROM instances WHERE name = ?`,
		name,
	)
	inst := &Instance{}
	var deletedAt sql.NullTime
	if err := row.Scan(
		&inst.ID, &inst.Name, &inst.Port, &inst.DBPath, &inst.LogPath,
		&inst.CreatedAt, &inst.UpdatedAt, &deletedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("instance %q not found", name)
		}
		return nil, fmt.Errorf("cannot query instance: %w", err)
	}
	if deletedAt.Valid {
		inst.DeletedAt = &deletedAt.Time
	}
	return inst, nil
}

// GetInstanceByID retrieves an instance by ID.
func (s *Store) GetInstanceByID(id int64) (*Instance, error) {
	row := s.db.QueryRow(
		`SELECT id, name, port, db_path, log_path, created_at, updated_at, deleted_at
		 FROM instances WHERE id = ?`,
		id,
	)
	inst := &Instance{}
	var deletedAt sql.NullTime
	if err := row.Scan(
		&inst.ID, &inst.Name, &inst.Port, &inst.DBPath, &inst.LogPath,
		&inst.CreatedAt, &inst.UpdatedAt, &deletedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("instance id %d not found", id)
		}
		return nil, fmt.Errorf("cannot query instance: %w", err)
	}
	if deletedAt.Valid {
		inst.DeletedAt = &deletedAt.Time
	}
	return inst, nil
}

// ListInstances returns all non-deleted instances, sorted by name.
func (s *Store) ListInstances() ([]*Instance, error) {
	rows, err := s.db.Query(
		`SELECT id, name, port, db_path, log_path, created_at, updated_at, deleted_at
		 FROM instances WHERE deleted_at IS NULL ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list instances: %w", err)
	}
	defer rows.Close()

	var instances []*Instance
	for rows.Next() {
		inst := &Instance{}
		var deletedAt sql.NullTime
		if err := rows.Scan(
			&inst.ID, &inst.Name, &inst.Port, &inst.DBPath, &inst.LogPath,
			&inst.CreatedAt, &inst.UpdatedAt, &deletedAt,
		); err != nil {
			return nil, fmt.Errorf("cannot scan instance: %w", err)
		}
		if deletedAt.Valid {
			inst.DeletedAt = &deletedAt.Time
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

// UpdateInstancePort updates the port for an instance.
func (s *Store) UpdateInstancePort(name string, newPort int) error {
	_, err := s.db.Exec(
		`UPDATE instances SET port = ?, updated_at = datetime('now') WHERE name = ?`,
		newPort, name,
	)
	if err != nil {
		return fmt.Errorf("cannot update port: %w", err)
	}
	return nil
}

// RenameInstance renames an instance.
func (s *Store) RenameInstance(oldName, newName string) error {
	_, err := s.db.Exec(
		`UPDATE instances SET name = ?, updated_at = datetime('now') WHERE name = ?`,
		newName, oldName,
	)
	if err != nil {
		return fmt.Errorf("cannot rename instance: %w", err)
	}
	return nil
}

// DeleteInstance soft-deletes an instance (sets deleted_at).
func (s *Store) DeleteInstance(name string) error {
	now := time.Now().UTC()
	_, err := s.db.Exec(
		`UPDATE instances SET deleted_at = ?, updated_at = ? WHERE name = ?`,
		now, now, name,
	)
	if err != nil {
		return fmt.Errorf("cannot delete instance: %w", err)
	}
	return nil
}

// HardDeleteInstance permanently removes an instance and its launch history.
func (s *Store) HardDeleteInstance(name string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM launch_history WHERE instance_id = (SELECT id FROM instances WHERE name = ?)`, name); err != nil {
		return fmt.Errorf("cannot delete launch history: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM instances WHERE name = ?`, name); err != nil {
		return fmt.Errorf("cannot delete instance: %w", err)
	}
	return tx.Commit()
}

// RecordLaunch inserts a new launch record and returns its ID.
func (s *Store) RecordLaunch(instanceID int64) (int64, error) {
	now := time.Now().UTC()
	result, err := s.db.Exec(
		`INSERT INTO launch_history (instance_id, start_time) VALUES (?, ?)`,
		instanceID, now,
	)
	if err != nil {
		return 0, fmt.Errorf("cannot record launch: %w", err)
	}
	id, _ := result.LastInsertId()
	return id, nil
}

// RecordKill updates the most recent launch record with end time and duration.
func (s *Store) RecordKill(instanceID int64) error {
	now := time.Now().UTC()
	// SQLite doesn't support ORDER BY ... LIMIT in UPDATE, so use a subquery
	_, err := s.db.Exec(
		`UPDATE launch_history
		 SET end_time = ?, duration = CAST(strftime('%s', ?) - strftime('%s', start_time) AS INTEGER)
		 WHERE id = (
			 SELECT id FROM launch_history
			 WHERE instance_id = ? AND end_time IS NULL
			 ORDER BY start_time DESC LIMIT 1
		 )`,
		now, now, instanceID,
	)
	if err != nil {
		return fmt.Errorf("cannot record kill: %w", err)
	}
	return nil
}

// GetLaunchHistory returns all launch records for an instance.
func (s *Store) GetLaunchHistory(instanceID int64) ([]*LaunchRecord, error) {
	rows, err := s.db.Query(
		`SELECT id, instance_id, start_time, end_time, duration
		 FROM launch_history WHERE instance_id = ? ORDER BY start_time`,
		instanceID,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot query launch history: %w", err)
	}
	defer rows.Close()

	var records []*LaunchRecord
	for rows.Next() {
		rec := &LaunchRecord{}
		var endTime sql.NullTime
		var duration sql.NullInt64
		if err := rows.Scan(&rec.ID, &rec.InstanceID, &rec.StartTime, &endTime, &duration); err != nil {
			return nil, fmt.Errorf("cannot scan launch record: %w", err)
		}
		if endTime.Valid {
			rec.EndTime = &endTime.Time
		}
		if duration.Valid {
			rec.DurationSec = &duration.Int64
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

// GetConfig retrieves a config value.
func (s *Store) GetConfig(key string) (string, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM config WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetConfig sets a config value.
func (s *Store) SetConfig(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO config (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}
