//go:build linux || darwin || freebsd || openbsd || netbsd

package proc

// forkArgs returns the extra mongod arguments needed to daemonize on Unix.
// On Windows mongod runs as a foreground service, so no fork flag is used.
func forkArgs() []string {
	return []string{"--fork"}
}

// forkedByMongod reports whether mongod daemonizes itself (Unix --fork).
// On Windows mongod stays attached to the launching process, so the caller
// must detach it explicitly via Start + Process.Release.
func forkedByMongod() bool {
	return true
}
