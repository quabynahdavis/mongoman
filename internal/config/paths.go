// Package config provides platform-aware directory paths for mongoman.
//
// Directory Layout (from plan.txt):
//
//	Data:     $HOME/mongoman/data/<name>
//	Logs:     $HOME/mongoman/logs/<name>.log
//	Backups:  $HOME/mongoman/backups/<name>_<ts>.tar.gz (or .zip on Windows)
//	Config:   ~/.config/mongoman/<name>.json —OR— %APPDATA%\mongoman\<name>.json
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	AppName = "mongoman"
)

// Paths holds all directory paths used by mongoman.
type Paths struct {
	// DataDir is the root for instance data directories.
	DataDir string
	// LogsDir is the root for instance log files.
	LogsDir string
	// BackupsDir is the root for backup archives.
	BackupsDir string
	// ConfigDir stores per-instance metadata JSON files.
	ConfigDir string
}

// DefaultPaths returns the platform-appropriate Paths, ensuring all
// directories exist.
func DefaultPaths() (*Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	mongomanDir := filepath.Join(home, AppName)

	var configDir string
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		configDir = filepath.Join(appData, AppName)
	default:
		// XDG_CONFIG_HOME or ~/.config
		xdg := os.Getenv("XDG_CONFIG_HOME")
		if xdg == "" {
			xdg = filepath.Join(home, ".config")
		}
		configDir = filepath.Join(xdg, AppName)
	}

	p := &Paths{
		DataDir:    filepath.Join(mongomanDir, "data"),
		LogsDir:    filepath.Join(mongomanDir, "logs"),
		BackupsDir: filepath.Join(mongomanDir, "backups"),
		ConfigDir:  configDir,
	}

	if err := p.ensureDirs(); err != nil {
		return nil, err
	}
	return p, nil
}

// ensureDirs creates all directories if they don't exist.
func (p *Paths) ensureDirs() error {
	for _, d := range []string{p.DataDir, p.LogsDir, p.BackupsDir, p.ConfigDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("cannot create directory %s: %w", d, err)
		}
	}
	return nil
}

// MetaFile returns the path to the JSON metadata file for instance name.
func (p *Paths) MetaFile(name string) string {
	return filepath.Join(p.ConfigDir, name+".json")
}

// DBPath returns the data directory for instance name.
func (p *Paths) DBPath(name string) string {
	return filepath.Join(p.DataDir, name)
}

// LogPath returns the log file path for instance name.
func (p *Paths) LogPath(name string) string {
	return filepath.Join(p.LogsDir, name+".log")
}

// BackupPath returns a timestamped backup file path for instance name.
func (p *Paths) BackupPath(name, timestamp string) string {
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	filename := fmt.Sprintf("%s_%s%s", name, timestamp, ext)
	return filepath.Join(p.BackupsDir, filename)
}
