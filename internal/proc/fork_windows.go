//go:build windows

package proc

// forkArgs returns the extra mongod arguments for daemonizing.
// On Windows mongod has no --fork flag, so none are added.
func forkArgs() []string {
	return nil
}

// forkedByMongod reports whether mongod daemonizes itself (Unix --fork).
// On Windows mongod stays attached to the launching process, so the caller
// must detach it explicitly via Start + Process.Release.
func forkedByMongod() bool {
	return false
}
