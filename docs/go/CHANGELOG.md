# Go Packages Changelog

Tracks changes to the Go source packages.

## [1.0.0] — 2026-05-20T04:27:00Z

### Added
- `internal/config/paths.go` — Platform-aware path resolution with XDG/AppData support
- `internal/instance/instance.go` — Instance CRUD with JSON metadata, launch history, rename/reconfigure/clone
- `internal/proc/proc.go` — mongod process launch/kill with pgrep-based PID lookup
- `internal/service/service.go` — Manager interface and Status type
- `internal/service/service_linux.go` — systemd service management
- `internal/service/service_darwin.go` — launchd service management
- `internal/service/service_windows.go` — Windows Service management via sc.exe
- `internal/service/service_stub.go` — BSD/fallback stub with error messages
- `main.go` — CLI entry point with 20 commands, argument dispatch, platform detection
