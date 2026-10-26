//go:build linux || darwin || freebsd || openbsd || netbsd

package proc

import (
	"os/exec"
	"strconv"
	"strings"
)

// findPIDByPortImpl locates a mongod process listening on the given port
// using pgrep to match the command line.
func findPIDByPortImpl(port string) (int, error) {
	cmd := exec.Command("pgrep", "-f", "mongod.*--port "+port)
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
