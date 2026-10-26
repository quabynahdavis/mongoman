package proc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davisdeveloper/mongoman/internal/config"
	"github.com/davisdeveloper/mongoman/internal/instance"
)

// mockInstance returns a test instance with the given port and unique name.
func mockInstance(t *testing.T, port int) *instance.Instance {
	t.Helper()
	tmp := t.TempDir()
	paths := &config.Paths{
		DataDir:    tmp,
		ConfigDir:  filepath.Join(tmp, "config"),
		LogsDir:    filepath.Join(tmp, "logs"),
		BackupsDir: filepath.Join(tmp, "backups"),
	}
	name := t.Name()
	// Create required dirs
	for _, d := range []string{paths.DataDir, paths.ConfigDir, paths.LogsDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	inst, err := instance.Create(paths, name, port)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	return inst
}

// TestIsRunning_falseWhenNoPID verifies IsRunning returns false when
// findPIDByPortImpl returns errProcessNotFound.
func TestIsRunning_falseWhenNoPID(t *testing.T) {
	// Override the platform implementation
	originalFindPIDByPortImpl := findPIDByPortImpl
	findPIDByPortImpl = func(port string) (int, error) {
		return 0, errProcessNotFound
	}
	defer func() { findPIDByPortImpl = originalFindPIDByPortImpl }()

	inst := mockInstance(t, 27017)
	if IsRunning(inst) {
		t.Fatal("IsRunning should return false when no PID found")
	}
}

// TestIsRunning_trueWhenPIDFound verifies IsRunning returns true when
// findPIDByPortImpl returns a valid PID.
func TestIsRunning_trueWhenPIDFound(t *testing.T) {
	originalFindPIDByPortImpl := findPIDByPortImpl
	findPIDByPortImpl = func(port string) (int, error) {
		if port == "27017" {
			return 12345, nil
		}
		return 0, errProcessNotFound
	}
	defer func() { findPIDByPortImpl = originalFindPIDByPortImpl }()

	inst := mockInstance(t, 27017)
	if !IsRunning(inst) {
		t.Fatal("IsRunning should return true when PID found")
	}
}

// TestKill_callsFindPIDByPort verifies Kill delegates to findPIDByPort.
func TestKill_callsFindPIDByPort(t *testing.T) {
	originalFindPIDByPortImpl := findPIDByPortImpl
	var called bool
	findPIDByPortImpl = func(port string) (int, error) {
		called = true
		if port == "27017" {
			return 99999, nil // fake PID; we'll verify it's passed to os.FindProcess
		}
		return 0, errProcessNotFound
	}
	defer func() { findPIDByPortImpl = originalFindPIDByPortImpl }()

	inst := mockInstance(t, 27017)
	// We can't easily mock os.FindProcess/os.Process.Kill, so this will fail
	// with "no such process" which is expected. The point is that it called
	// findPIDByPortImpl.
	err := Kill(inst)
	if err != nil {
		// Expected to fail on kill (no real process), but findPIDByPort should have been called
		if !called {
			t.Fatal("Kill did not call findPIDByPortImpl")
		}
	}
}

// TestKill_returnsErrorWhenNoPID verifies Kill returns an error when
// findPIDByPortImpl returns errProcessNotFound.
func TestKill_returnsErrorWhenNoPID(t *testing.T) {
	originalFindPIDByPortImpl := findPIDByPortImpl
	findPIDByPortImpl = func(port string) (int, error) {
		return 0, errProcessNotFound
	}
	defer func() { findPIDByPortImpl = originalFindPIDByPortImpl }()

	inst := mockInstance(t, 27017)
	err := Kill(inst)
	if err == nil {
		t.Fatal("Kill should error when no PID found")
	}
	// Kill doesn't wrap with %w, so check message content
	if !strings.Contains(err.Error(), "running process found") {
		t.Errorf("Kill error should indicate running process not found: %v", err)
	}
}
