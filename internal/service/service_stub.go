//go:build !linux && !darwin && !windows

package service

import (
	"fmt"
	"os"
	"os/exec"
)

type stubManager struct{}

func newPlatformManager() Manager {
	return &stubManager{}
}

func (m *stubManager) Enable(name string, port int, dbPath, logPath string) error {
	fmt.Fprintf(os.Stderr, "⚠️  OS service management is not yet supported on this platform.\n")
	fmt.Fprintf(os.Stderr, "   You can still run instances directly with 'mongoman launch %s'\n", name)
	fmt.Fprintf(os.Stderr, "   On BSD, consider adding a systemd unit or rc.d script manually.\n")
	return &errNotSupported{"systemd/launchd"}
}

func (m *stubManager) Disable(name string) error {
	return &errNotSupported{"systemd/launchd"}
}

func (m *stubManager) Start(name string) error {
	return &errNotSupported{"systemd/launchd"}
}

func (m *stubManager) Stop(name string) error {
	return &errNotSupported{"systemd/launchd"}
}

func (m *stubManager) Restart(name string) error {
	return &errNotSupported{"systemd/launchd"}
}

func (m *stubManager) Status(name string) Status {
	// Fall back to checking if a direct process is running.
	out, err := exec.Command("pgrep", "-f", fmt.Sprintf("mongod.*%s", name)).Output()
	if err == nil && len(out) > 0 {
		return StatusRunning
	}
	return StatusStopped
}

func (m *stubManager) IsEnabled(name string) bool {
	return false
}

var _ Manager = (*stubManager)(nil)
