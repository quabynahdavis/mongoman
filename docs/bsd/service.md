# BSD Service Management — Stub

## Overview

BSD platforms (FreeBSD, OpenBSD, NetBSD) don't have a unified service manager that matches systemd or launchd. The `stubManager` in `service_stub.go` provides a fallback that:

1. Prints a helpful message about manual service configuration
2. Supports direct process management (launch/kill) via the `proc` package
3. Provides basic status detection via `pgrep`

## Stub Behavior

| Method | Behavior |
|--------|----------|
| `Enable` | Prints message: "OS service management is not yet supported on this platform. You can still run instances directly with 'mongoman launch <name>'. On BSD, consider adding a systemd unit or rc.d script manually." |
| `Disable` | Returns "not supported" error |
| `Start` | Returns "not supported" error |
| `Stop` | Returns "not supported" error |
| `Restart` | Returns "not supported" error |
| `Status` | Falls back to `pgrep -f "mongod.*<name>"` to check if direct process is running |
| `IsEnabled` | Always returns `false` |

## Future Work

BSD service management could be added by:
- **FreeBSD**: rc.d scripts in `/usr/local/etc/rc.d/`
- **OpenBSD**: rc.conf.d configuration
- **NetBSD**: rc.d scripts

Each would need a new build-tagged file (e.g., `service_freebsd.go` with `//go:build freebsd`).
