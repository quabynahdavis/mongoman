// Package instance manages MongoDB instance metadata and CRUD operations.
package instance

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/davisdeveloper/mongoman/internal/config"
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
	Meta     Metadata
	Paths    *config.Paths
	MetaPath string
	DBPath   string
	LogPath  string
}

// Load reads metadata for a named instance. Returns nil if it does not exist.
func Load(paths *config.Paths, name string) (*Instance, error) {
	metaPath := paths.MetaFile(name)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("instance %q not found", name)
		}
		return nil, fmt.Errorf("cannot read metadata for %q: %w", name, err)
	}

	var m Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("invalid metadata for %q: %w", name, err)
	}

	return &Instance{
		Meta:     m,
		Paths:    paths,
		MetaPath: metaPath,
		DBPath:   paths.DBPath(name),
		LogPath:  paths.LogPath(name),
	}, nil
}

// Create initializes a new instance with the given name and port.
func Create(paths *config.Paths, name string, port int) (*Instance, error) {
	metaPath := paths.MetaFile(name)

	// Check for conflict.
	if _, err := os.Stat(metaPath); err == nil {
		return nil, fmt.Errorf("instance %q already exists", name)
	}

	now := time.Now()
	inst := &Instance{
		Meta: Metadata{
			Name:      name,
			Port:      port,
			CreatedAt: now,
		},
		Paths:    paths,
		MetaPath: metaPath,
		DBPath:   paths.DBPath(name),
		LogPath:  paths.LogPath(name),
	}

	// Create data directory and log file.
	if err := os.MkdirAll(inst.DBPath, 0o755); err != nil {
		return nil, fmt.Errorf("cannot create data directory: %w", err)
	}
	if err := os.WriteFile(inst.LogPath, []byte{}, 0o644); err != nil {
		return nil, fmt.Errorf("cannot create log file: %w", err)
	}

	// Write metadata.
	if err := inst.save(); err != nil {
		return nil, err
	}

	return inst, nil
}

// Exists checks whether an instance's metadata file exists.
func Exists(paths *config.Paths, name string) bool {
	_, err := os.Stat(paths.MetaFile(name))
	return err == nil
}

// RequireExists loads an instance, returning an error if it is missing.
// It is a thin wrapper around Load for callers that want a concise API.
func RequireExists(paths *config.Paths, name string) (*Instance, error) {
	return Load(paths, name)
}

// save writes the metadata JSON to disk.
func (inst *Instance) save() error {
	data, err := json.MarshalIndent(inst.Meta, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal metadata: %w", err)
	}
	if err := os.WriteFile(inst.MetaPath, data, 0o644); err != nil {
		return fmt.Errorf("cannot write metadata: %w", err)
	}
	return nil
}

// RecordLaunch adds a launch record and increments the counter.
func (inst *Instance) RecordLaunch() error {
	inst.Meta.LaunchCount++
	inst.Meta.LaunchHistory = append(inst.Meta.LaunchHistory, LaunchRecord{
		Start: time.Now(),
	})
	return inst.save()
}

// RecordKill sets the end time on the most recent launch record.
func (inst *Instance) RecordKill() error {
	if len(inst.Meta.LaunchHistory) == 0 {
		return nil
	}
	now := time.Now()
	inst.Meta.LaunchHistory[len(inst.Meta.LaunchHistory)-1].End = &now
	return inst.save()
}

// MarkDeleted sets the deletion timestamp.
func (inst *Instance) MarkDeleted() error {
	now := time.Now()
	inst.Meta.DeletedAt = &now
	return inst.save()
}

// Rename changes the instance name and moves metadata & data on disk.
func Rename(paths *config.Paths, oldName, newName string) error {
	inst, err := RequireExists(paths, oldName)
	if err != nil {
		return err
	}

	// Move metadata file.
	oldMeta := paths.MetaFile(oldName)
	newMeta := paths.MetaFile(newName)
	if err := os.Rename(oldMeta, newMeta); err != nil {
		return fmt.Errorf("cannot rename metadata: %w", err)
	}

	// Move data directory.
	oldDB := paths.DBPath(oldName)
	newDB := paths.DBPath(newName)
	if err := os.Rename(oldDB, newDB); err != nil {
		return fmt.Errorf("cannot rename data directory: %w", err)
	}

	// Update the metadata in-place.
	inst.Meta.Name = newName
	inst.MetaPath = newMeta
	inst.DBPath = newDB
	inst.LogPath = paths.LogPath(newName)
	return inst.save()
}

// Reconfigure updates the port number for an instance.
func Reconfigure(paths *config.Paths, name string, newPort int) error {
	inst, err := RequireExists(paths, name)
	if err != nil {
		return err
	}
	inst.Meta.Port = newPort
	return inst.save()
}

// Clone copies an instance's data and creates new metadata.
func Clone(paths *config.Paths, srcName, dstName string, port int) error {
	if _, err := Load(paths, srcName); err != nil {
		return fmt.Errorf("source instance %q not found: %w", srcName, err)
	}

	srcDB := paths.DBPath(srcName)
	dstDB := paths.DBPath(dstName)

	// Copy data directory using a pure-Go recursive copy (cross-platform).
	if err := copyDir(srcDB, dstDB); err != nil {
		return fmt.Errorf("cannot clone data directory: %w", err)
	}

	// Create new metadata.
	_, err := Create(paths, dstName, port)
	return err
}

// ListAll returns all instance names found in the config directory.
func ListAll(paths *config.Paths) ([]string, error) {
	entries, err := os.ReadDir(paths.ConfigDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read config directory: %w", err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && len(e.Name()) > 5 && e.Name()[len(e.Name())-5:] == ".json" {
			names = append(names, e.Name()[:len(e.Name())-5])
		}
	}
	return names, nil
}

// Delete removes instance data and optionally marks metadata as deleted.
func Delete(paths *config.Paths, name string) error {
	inst, err := RequireExists(paths, name)
	if err != nil {
		return err
	}

	// Remove data directory.
	if err := os.RemoveAll(inst.DBPath); err != nil {
		return fmt.Errorf("cannot remove data directory: %w", err)
	}

	// Remove metadata file (optional: keep for history).
	if err := os.Remove(inst.MetaPath); err != nil {
		return fmt.Errorf("cannot remove metadata: %w", err)
	}

	return nil
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
