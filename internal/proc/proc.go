// Package proc manages mongod processes — launching, killing, and checking status.
package proc

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/davisdeveloper/mongoman/internal/instance"
)

// Launch starts a mongod process for the given instance as a forked daemon.
func Launch(inst *instance.Instance) error {
	port := strconv.Itoa(inst.Meta.Port)

	args := []string{
		"--port", port,
		"--dbpath", inst.DBPath,
		"--bind_ip", "localhost",
		"--logpath", inst.LogPath,
	}

	// --fork is only available on Unix; on Windows we use START /B semantics.
	// We let mongod decide; if --fork fails, the user sees the error.
	args = append(args, "--fork")

	cmd := exec.Command("mongod", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch mongod for %q: %w", inst.Meta.Name, err)
	}

	return inst.RecordLaunch()
}

// Kill terminates the mongod process for the given instance by matching its port.
func Kill(inst *instance.Instance) error {
	port := strconv.Itoa(inst.Meta.Port)

	// Find the PID of the mongod process running on this port.
	pid, err := findPIDByPort(port)
	if err != nil {
		return fmt.Errorf("no running process found for %q on port %s", inst.Meta.Name, port)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("cannot find process %d: %w", pid, err)
	}

	if err := proc.Kill(); err != nil {
		return fmt.Errorf("failed to kill process %d: %w", pid, err)
	}

	fmt.Printf("Killed MongoDB process for %q (PID %d)\n", inst.Meta.Name, pid)
	return inst.RecordKill()
}

// IsRunning checks whether a mongod process is running on the instance's port.
func IsRunning(inst *instance.Instance) bool {
	port := strconv.Itoa(inst.Meta.Port)
	_, err := findPIDByPort(port)
	return err == nil
}

// findPIDByPort searches for a mongod process listening on the given port.
// Uses pgrep on Unix, Get-Process on Windows.
func findPIDByPort(port string) (int, error) {
	// Strategy: use pgrep to find mongod processes, then filter by --port argument.
	cmd := exec.Command("pgrep", "-f", fmt.Sprintf("mongod.*--port %s", port))
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("process not found")
	}

	lines := strings.Fields(string(output))
	if len(lines) == 0 {
		return 0, fmt.Errorf("process not found")
	}

	pid, err := strconv.Atoi(lines[0])
	if err != nil {
		return 0, fmt.Errorf("invalid PID: %w", err)
	}
	return pid, nil
}
