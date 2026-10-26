//go:build windows

package proc

import (
	"os/exec"
	"strconv"
	"strings"
)

func init() {
	findPIDByPortImpl = findPIDByPortWindows
}

// findPIDByPortWindows locates a mongod process listening on the given port
// by querying Win32_Process via PowerShell and matching the command line.
func findPIDByPortWindows(port string) (int, error) {
	// Match mongod processes whose command line contains --port <port>.
	cmd := exec.Command(
		"powershell",
		"-NoProfile",
		"-Command",
		"Get-CimInstance Win32_Process | Where-Object { $_.CommandLine -match 'mongod.*--port "+port+"' } | Select-Object -ExpandProperty ProcessId",
	)
	output, err := cmd.Output()
	if err != nil {
		return 0, errProcessNotFound
	}

	lines := strings.Fields(string(output))
	if len(lines) == 0 {
		return 0, errProcessNotFound
	}

	pid, err := strconv.Atoi(lines[0])
	if err != nil {
		return 0, err
	}
	return pid, nil
}
