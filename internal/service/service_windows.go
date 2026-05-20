//go:build windows

package service

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type windowsManager struct{}

func newPlatformManager() Manager {
	return &windowsManager{}
}

func (m *windowsManager) serviceName(name string) string {
	return fmt.Sprintf("MongoDB-%s", name)
}

func (m *windowsManager) Enable(name string, port int, dbPath, logPath string) error {
	svcName := m.serviceName(name)

	binaryPath := fmt.Sprintf(
		`"mongod" --port %d --dbpath "%s" --bind_ip localhost --logpath "%s"`,
		port, dbPath, logPath,
	)

	// Create the service using sc.exe (Windows SDK tool).
	args := []string{
		"create", svcName,
		"binPath=", binaryPath,
		"start=", "auto",
		"DisplayName=", fmt.Sprintf("MongoDB Instance %s", name),
	}

	cmd := exec.Command("sc.exe", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create Windows service: %w", err)
	}

	// Start the service.
	if err := exec.Command("sc.exe", "start", svcName).Run(); err != nil {
		return fmt.Errorf("failed to start Windows service: %w", err)
	}

	fmt.Printf("✅ Enabled and started Windows service %s (port %d)\n", svcName, port)
	return nil
}

func (m *windowsManager) Disable(name string) error {
	svcName := m.serviceName(name)

	// Stop the service.
	exec.Command("sc.exe", "stop", svcName).Run()

	// Delete the service.
	if err := exec.Command("sc.exe", "delete", svcName).Run(); err != nil {
		return fmt.Errorf("failed to delete Windows service: %w", err)
	}

	fmt.Printf("Disabled Windows service %s\n", svcName)
	return nil
}

func (m *windowsManager) Start(name string) error {
	return exec.Command("sc.exe", "start", m.serviceName(name)).Run()
}

func (m *windowsManager) Stop(name string) error {
	return exec.Command("sc.exe", "stop", m.serviceName(name)).Run()
}

func (m *windowsManager) Restart(name string) error {
	svcName := m.serviceName(name)
	exec.Command("sc.exe", "stop", svcName).Run()
	return exec.Command("sc.exe", "start", svcName).Run()
}

func (m *windowsManager) Status(name string) Status {
	svcName := m.serviceName(name)
	out, err := exec.Command("sc.exe", "query", svcName).Output()
	if err != nil {
		return StatusStopped
	}

	output := string(out)
	if strings.Contains(output, "RUNNING") {
		return StatusRunning
	}
	if strings.Contains(output, "STOPPED") {
		return StatusStopped
	}

	// Check if service config exists at all.
	err = exec.Command("sc.exe", "qc", svcName).Run()
	if err == nil {
		return StatusDisabled
	}

	return StatusStopped
}

func (m *windowsManager) IsEnabled(name string) bool {
	return exec.Command("sc.exe", "qc", m.serviceName(name)).Run() == nil
}

var _ Manager = (*windowsManager)(nil)
