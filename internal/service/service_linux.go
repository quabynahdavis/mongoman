//go:build linux

package service

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type systemdManager struct{}

func newPlatformManager() Manager {
	return &systemdManager{}
}

func (m *systemdManager) serviceName(name string) string {
	return fmt.Sprintf("mongod-%s", name)
}

func (m *systemdManager) servicePath(name string) string {
	return fmt.Sprintf("/etc/systemd/system/%s.service", m.serviceName(name))
}

func (m *systemdManager) Enable(name string, port int, dbPath, logPath string) error {
	svcName := m.serviceName(name)
	unit := fmt.Sprintf(`[Unit]
Description=MongoDB Instance - %s
After=network.target

[Service]
Type=forking
ExecStart=/usr/bin/mongod --port %d --dbpath %s --bind_ip localhost --logpath %s --fork
ExecStop=/usr/bin/mongod --shutdown --dbpath %s
PIDFile=%s/mongod.pid
User=%s
Restart=on-failure

[Install]
WantedBy=multi-user.target
`, name, port, dbPath, logPath, dbPath, dbPath, os.Getenv("USER"))

	path := m.servicePath(name)
	if err := os.WriteFile(path, []byte(unit), 0o644); err != nil {
		return fmt.Errorf("cannot write systemd unit file: %w", err)
	}

	// Reload systemd and enable.
	cmds := [][]string{
		{"systemctl", "daemon-reload"},
		{"systemctl", "enable", svcName},
		{"systemctl", "start", svcName},
	}
	for _, args := range cmds {
		if err := exec.Command(args[0], args[1:]...).Run(); err != nil {
			return fmt.Errorf("systemctl %s failed: %w", strings.Join(args[1:], " "), err)
		}
	}

	fmt.Printf("✅ Enabled and started systemd service %s (port %d)\n", svcName, port)
	return nil
}

func (m *systemdManager) Disable(name string) error {
	svcName := m.serviceName(name)

	// Stop and disable.
	exec.Command("systemctl", "stop", svcName).Run()
	exec.Command("systemctl", "disable", svcName).Run()

	// Remove unit file.
	os.Remove(m.servicePath(name))
	exec.Command("systemctl", "daemon-reload").Run()

	fmt.Printf("Disabled systemd service %s\n", svcName)
	return nil
}

func (m *systemdManager) Start(name string) error {
	svcName := m.serviceName(name)
	return exec.Command("systemctl", "start", svcName).Run()
}

func (m *systemdManager) Stop(name string) error {
	svcName := m.serviceName(name)
	return exec.Command("systemctl", "stop", svcName).Run()
}

func (m *systemdManager) Restart(name string) error {
	svcName := m.serviceName(name)
	return exec.Command("systemctl", "restart", svcName).Run()
}

func (m *systemdManager) Status(name string) Status {
	svcName := m.serviceName(name)

	// Check if running.
	if err := exec.Command("systemctl", "is-active", "--quiet", svcName).Run(); err == nil {
		return StatusRunning
	}

	// Check if enabled.
	if err := exec.Command("systemctl", "is-enabled", "--quiet", svcName).Run(); err == nil {
		return StatusEnabled
	}

	// Check if service file exists at all.
	if _, err := os.Stat(m.servicePath(name)); err == nil {
		return StatusDisabled
	}

	return StatusStopped
}

func (m *systemdManager) IsEnabled(name string) bool {
	svcName := m.serviceName(name)
	return exec.Command("systemctl", "is-enabled", "--quiet", svcName).Run() == nil
}

// getUID returns the numeric user ID (helper for systemd unit).
func getUID() string {
	return strconv.Itoa(os.Getuid())
}

// Ensure compatibility with the interface.
var _ Manager = (*systemdManager)(nil)

