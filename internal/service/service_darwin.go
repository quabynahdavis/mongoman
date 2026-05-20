//go:build darwin

package service

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

type launchdManager struct{}

func newPlatformManager() Manager {
	return &launchdManager{}
}

func (m *launchdManager) serviceName(name string) string {
	return fmt.Sprintf("com.mongoman.%s", name)
}

func (m *launchdManager) plistPath(name string) string {
	// LaunchAgents in user's Library directory.
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", m.serviceName(name)+".plist")
}

func (m *launchdManager) Enable(name string, port int, dbPath, logPath string) error {
	svcName := m.serviceName(name)
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/mongod</string>
        <string>--port</string>
        <string>%d</string>
        <string>--dbpath</string>
        <string>%s</string>
        <string>--bind_ip</string>
        <string>localhost</string>
        <string>--logpath</string>
        <string>%s</string>
        <string>--fork</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>%s</string>
    <key>StandardErrorPath</key>
    <string>%s</string>
</dict>
</plist>
`, svcName, port, dbPath, logPath, logPath, logPath)

	path := m.plistPath(name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("cannot create LaunchAgents directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(plist), 0o644); err != nil {
		return fmt.Errorf("cannot write plist file: %w", err)
	}

	// Load the launchd job.
	if err := exec.Command("launchctl", "load", path).Run(); err != nil {
		return fmt.Errorf("launchctl load failed: %w", err)
	}

	fmt.Printf("✅ Enabled and started launchd service %s (port %d)\n", svcName, port)
	return nil
}

func (m *launchdManager) Disable(name string) error {
	svcName := m.serviceName(name)
	path := m.plistPath(name)

	// Unload and remove.
	exec.Command("launchctl", "unload", path).Run()
	os.Remove(path)

	fmt.Printf("Disabled launchd service %s\n", svcName)
	return nil
}

func (m *launchdManager) Start(name string) error {
	return exec.Command("launchctl", "load", m.plistPath(name)).Run()
}

func (m *launchdManager) Stop(name string) error {
	return exec.Command("launchctl", "unload", m.plistPath(name)).Run()
}

func (m *launchdManager) Restart(name string) error {
	path := m.plistPath(name)
	exec.Command("launchctl", "unload", path).Run()
	return exec.Command("launchctl", "load", path).Run()
}

func (m *launchdManager) Status(name string) Status {
	svcName := m.serviceName(name)

	// launchctl list returns the PID if running, or exit code 0 even if not.
	out, err := exec.Command("launchctl", "list", svcName).Output()
	if err != nil {
		return StatusStopped
	}

	// The output format is: "PID	Status	Label"
	// If PID is "-", the process is not running.
	pid := string(out)
	if len(pid) > 0 && pid[0] != '-' && pid[0] != '\n' {
		return StatusRunning
	}

	// Check if plist exists (enabled).
	if _, err := os.Stat(m.plistPath(name)); err == nil {
		return StatusEnabled
	}

	return StatusStopped
}

func (m *launchdManager) IsEnabled(name string) bool {
	_, err := os.Stat(m.plistPath(name))
	return err == nil
}

var _ Manager = (*launchdManager)(nil)

func init() {
	// Ensure ~/Library/LaunchAgents exists.
	home, _ := os.UserHomeDir()
	os.MkdirAll(filepath.Join(home, "Library", "LaunchAgents"), 0o755)
}

// Import used by build tag but referenced indirectly.
var _ = user.Current
