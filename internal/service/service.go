// Package service provides OS-native service management for mongod instances.
//
// Supported backends:
//   - Linux:   systemd
//   - macOS:   launchd
//   - Windows: Windows Service (sc.exe / New-Service)
//   - BSD:     rc.d / systemd (if available)
package service

import (
	"fmt"
)

// Status represents the current service state.
type Status int

const (
	StatusUnknown Status = iota
	StatusRunning
	StatusStopped
	StatusEnabled
	StatusDisabled
)

func (s Status) String() string {
	switch s {
	case StatusRunning:
		return "✅ Running"
	case StatusStopped:
		return "❌ Stopped"
	case StatusEnabled:
		return "🔁 Enabled"
	case StatusDisabled:
		return "⛔ Disabled"
	default:
		return "❓ Unknown"
	}
}

// Manager handles OS service lifecycle for a mongod instance.
type Manager interface {
	// Enable creates and starts the OS service for the instance.
	Enable(name string, port int, dbPath, logPath string) error
	// Disable stops and removes the OS service.
	Disable(name string) error
	// Start starts the OS service.
	Start(name string) error
	// Stop stops the OS service.
	Stop(name string) error
	// Restart restarts the OS service.
	Restart(name string) error
	// Status returns the current service status.
	Status(name string) Status
	// IsEnabled reports whether the service exists and is enabled.
	IsEnabled(name string) bool
}

// NewManager returns the platform-appropriate service manager.
func NewManager() Manager {
	return newPlatformManager()
}

// errNotSupported is used by stub platforms.
type errNotSupported struct {
	feature string
}

func (e *errNotSupported) Error() string {
	return fmt.Sprintf("OS service management (%s) is not supported on this platform", e.feature)
}
