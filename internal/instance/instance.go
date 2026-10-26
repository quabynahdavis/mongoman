// Package instance manages MongoDB instance metadata and CRUD operations.
// This is a thin wrapper around the SQLite store for backward compatibility.
package instance

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/quabynahdavis/mongoman/internal/config"
	"github.com/quabynahdavis/mongoman/internal/store/sqlite"
)

// LaunchRecord tracks a single launch session.
type LaunchRecord struct {
	Start time.Time  `json:"start"`
	End   *time.Time `json:"end,omitempty"`
}

// Metadata represents the JSON metadata stored for each instance.
type Metadata struct {
	Name          string         `json:"name"`
	Port          int            `json:"port"`
	CreatedAt     time.Time      `json:"created_at"`
	LaunchCount   int            `json:"launch_count"`
	LaunchHistory []LaunchRecord `json:"launch_history,omitempty"`
	DeletedAt     *time.Time     `json:"deleted_at,omitempty"`
}

// Instance wraps Metadata together with filesystem paths.
type Instance struct {
	ID        int64
	Name      string
	Port      int
	DBPath    string
	LogPath   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	// Meta provides backward compatibility with the old JSON-based API.
	Meta Metadata
}

// store holds the SQLite store instance.
var store *sqlite.Store

// storeConfigDir tracks the config directory used to initialize the store.
var storeConfigDir string

// initStore initializes the SQLite store if not already done,
// or re-initializes if the config directory has changed (for tests).
func initStore(paths *config.Paths) error {
	if store != nil && storeConfigDir == paths.ConfigDir {
		return nil
	}
	var err error
	store, err = sqlite.NewStore(paths)
	if err != nil {
		return err
	}
	storeConfigDir = paths.ConfigDir
	return nil
}

// Load reads metadata for a named instance. Returns nil if it does not exist.
func Load(paths *config.Paths, name string) (*Instance, error) {
	if err := initStore(paths); err != nil {
		return nil, err
	}

	inst, err := store.GetInstanceByName(name)
	if err != nil {
		return nil, err
	}

	// Fetch launch history for Meta.LaunchCount and Meta.LaunchHistory
	history, _ := store.GetLaunchHistory(inst.ID)
	launchHistory := make([]LaunchRecord, len(history))
	for i, h := range history {
		launchHistory[i] = LaunchRecord{
			Start: h.StartTime,
			End:   h.EndTime,
		}
	}

	return &Instance{
		ID:        inst.ID,
		Name:      inst.Name,
		Port:      inst.Port,
		DBPath:    inst.DBPath,
		LogPath:   inst.LogPath,
		CreatedAt: inst.CreatedAt,
		UpdatedAt: inst.UpdatedAt,
		DeletedAt: inst.DeletedAt,
		Meta: Metadata{
			Name:          inst.Name,
			Port:          inst.Port,
			CreatedAt:     inst.CreatedAt,
			LaunchCount:   len(history),
			LaunchHistory: launchHistory,
			DeletedAt:     inst.DeletedAt,
		},
	}, nil
}

// Create initializes a new instance with the given name and port.
func Create(paths *config.Paths, name string, port int) (*Instance, error) {
	if err := initStore(paths); err != nil {
		return nil, err
	}

	// Check for conflict.
	if _, err := store.GetInstanceByName(name); err == nil {
		return nil, fmt.Errorf("instance %q already exists", name)
	}

	dbPath := paths.DBPath(name)
	logPath := paths.LogPath(name)

	// Create data directory and log file.
	if err := os.MkdirAll(dbPath, 0o755); err != nil {
		return nil, fmt.Errorf("cannot create data directory: %w", err)
	}
	if err := os.WriteFile(logPath, []byte{}, 0o644); err != nil {
		return nil, fmt.Errorf("cannot create log file: %w", err)
	}

	id, err := store.CreateInstance(name, port, dbPath, logPath)
	if err != nil {
		return nil, err
	}

	inst, err := store.GetInstanceByID(id)
	if err != nil {
		return nil, err
	}

	// Fetch launch history for Meta.LaunchCount and Meta.LaunchHistory
	history, _ := store.GetLaunchHistory(inst.ID)
	launchHistory := make([]LaunchRecord, len(history))
	for i, h := range history {
		launchHistory[i] = LaunchRecord{
			Start: h.StartTime,
			End:   h.EndTime,
		}
	}

	return &Instance{
		ID:        inst.ID,
		Name:      inst.Name,
		Port:      inst.Port,
		DBPath:    inst.DBPath,
		LogPath:   inst.LogPath,
		CreatedAt: inst.CreatedAt,
		UpdatedAt: inst.UpdatedAt,
		DeletedAt: inst.DeletedAt,
		Meta: Metadata{
			Name:          inst.Name,
			Port:          inst.Port,
			CreatedAt:     inst.CreatedAt,
			LaunchCount:   len(history),
			LaunchHistory: launchHistory,
			DeletedAt:     inst.DeletedAt,
		},
	}, nil
}

// Exists checks whether an instance's metadata exists in the database (not soft-deleted).
func Exists(paths *config.Paths, name string) bool {
	if err := initStore(paths); err != nil {
		return false
	}
	inst, err := store.GetInstanceByName(name)
	if err != nil {
		return false
	}
	return inst.DeletedAt == nil
}

// RequireExists loads an instance, returning an error if it is missing.
func RequireExists(paths *config.Paths, name string) (*Instance, error) {
	return Load(paths, name)
}

// RecordLaunch adds a launch record and increments the counter.
func (inst *Instance) RecordLaunch() error {
	if store == nil {
		return fmt.Errorf("store not initialized")
	}
	id, err := store.RecordLaunch(inst.ID)
	if err != nil {
		return err
	}
	inst.ID = id // refresh in case
	return nil
}

// RecordKill sets the end time on the most recent launch record.
func (inst *Instance) RecordKill() error {
	if store == nil {
		return fmt.Errorf("store not initialized")
	}
	return store.RecordKill(inst.ID)
}

// MarkDeleted sets the deletion timestamp.
func (inst *Instance) MarkDeleted() error {
	if store == nil {
		return fmt.Errorf("store not initialized")
	}
	now := time.Now()
	inst.DeletedAt = &now
	return nil // Store handles this via DeleteInstance
}

// Rename changes the instance name and moves metadata & data on disk.
func Rename(paths *config.Paths, oldName, newName string) error {
	if err := initStore(paths); err != nil {
		return err
	}

	_, err := Load(paths, oldName)
	if err != nil {
		return err
	}

	// Move data directory.
	oldDB := paths.DBPath(oldName)
	newDB := paths.DBPath(newName)
	if err := os.Rename(oldDB, newDB); err != nil {
		return fmt.Errorf("cannot rename data directory: %w", err)
	}

	// Update the database record.
	if err := store.RenameInstance(oldName, newName); err != nil {
		return err
	}

	return nil
}

// Reconfigure updates the port number for an instance.
func Reconfigure(paths *config.Paths, name string, newPort int) error {
	if err := initStore(paths); err != nil {
		return err
	}

	// Verify instance exists
	if _, err := Load(paths, name); err != nil {
		return err
	}

	return store.UpdateInstancePort(name, newPort)
}

// Clone copies an instance's data and creates new metadata.
func Clone(paths *config.Paths, srcName, dstName string, port int) error {
	if err := initStore(paths); err != nil {
		return err
	}

	_, err := Load(paths, srcName)
	if err != nil {
		return fmt.Errorf("source instance %q not found: %w", srcName, err)
	}

	srcDB := paths.DBPath(srcName)
	dstDB := paths.DBPath(dstName)

	// Copy data directory using a pure-Go recursive copy (cross-platform).
	if err := copyDir(srcDB, dstDB); err != nil {
		return fmt.Errorf("cannot clone data directory: %w", err)
	}

	// Create new metadata in SQLite.
	_, err = Create(paths, dstName, port)
	return err
}

// ListAll returns all instance names found in the database.
func ListAll(paths *config.Paths) ([]string, error) {
	if err := initStore(paths); err != nil {
		return nil, err
	}

	instances, err := store.ListInstances()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(instances))
	for i, inst := range instances {
		names[i] = inst.Name
	}
	return names, nil
}

// Delete removes instance data and marks metadata as deleted.
func Delete(paths *config.Paths, name string) error {
	if err := initStore(paths); err != nil {
		return err
	}

	inst, err := Load(paths, name)
	if err != nil {
		return err
	}

	// Remove data directory.
	if err := os.RemoveAll(inst.DBPath); err != nil {
		return fmt.Errorf("cannot remove data directory: %w", err)
	}

	// Soft delete in database.
	return store.DeleteInstance(name)
}

// copyDir recursively copies a directory from src to dst using pure Go,
// preserving file modes. Cross-platform — no shell commands required.
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("cannot stat source %s: %w", src, err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source %q is not a directory", src)
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("cannot create destination %s: %w", dst, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("cannot read source directory %s: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("cannot copy %s: %w", srcPath, err)
		}
	}
	return nil
}

// copyFile copies a single regular file from src to dst, preserving mode.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("cannot open source %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("cannot create destination %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("cannot copy contents of %s: %w", src, err)
	}

	// Preserve source file mode.
	if info, err := os.Stat(src); err == nil {
		if chmodErr := out.Chmod(info.Mode()); chmodErr != nil {
			// Non-fatal: mode may not be settable on all filesystems.
			_ = chmodErr
		}
	}
	return out.Sync()
}
