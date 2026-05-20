# Architecture Overview

## mongoman — Cross-Platform MongoDB Instance Manager

### Purpose

mongoman manages multiple MongoDB instances on a single machine. Each instance has its own port, data directory, log file, and metadata. Instances can run as direct forked processes or as OS-native services (systemd, launchd, Windows Service).

### Design Principles

1. **Single binary** — No runtime dependencies. Written in Go, produces a statically-linked executable.
2. **Cross-platform** — Same codebase targets Linux, macOS, BSD, and Windows.
3. **Portable paths** — Follows XDG Base Directory spec on Unix, `%APPDATA%` on Windows.
4. **JSON metadata** — All instance state stored in human-readable JSON files.
5. **Zero external dependencies** — Pure Go standard library only.

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        main.go                              │
│  CLI argument parsing, command dispatch, error handling     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  instance/    │  │  proc/       │  │  service/        │  │
│  │  CRUD ops     │  │  launch/kill │  │  OS svc mgmt     │  │
│  │  JSON meta    │  │  PID lookup  │  │  systemd/launchd │  │
│  └──────────────┘  └──────────────┘  │  /Windows svc     │  │
│                                       └──────────────────┘  │
│                                                             │
│  ┌──────────────┐                                           │
│  │  config/     │                                           │
│  │  path mgmt   │                                           │
│  └──────────────┘                                           │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

1. User runs `mongoman <command> <args>`
2. `main.go` parses `os.Args`, dispatches to the correct handler
3. Handler calls into `instance`, `proc`, or `service` packages
4. All packages use `config.Paths` for directory resolution
5. Results printed to stdout; errors to stderr with exit code 1

### Directory Layout (from plan.txt)

| Content | Unix Path | Windows Path |
|---------|-----------|--------------|
| Data | `~/mongoman/data/<name>` | `%USERPROFILE%\mongoman\data\<name>` |
| Logs | `~/mongoman/logs/<name>.log` | `%USERPROFILE%\mongoman\logs\<name>.log` |
| Backups | `~/mongoman/backups/` | `%USERPROFILE%\mongoman\backups\` |
| Config | `~/.config/mongoman/<name>.json` | `%APPDATA%\mongoman\<name>.json` |

### Instance Metadata Format

Each instance stores a JSON file with its metadata:

```json
{
  "name": "dev27018",
  "port": 27018,
  "created_at": "2026-05-20T04:12:47Z",
  "launch_count": 3,
  "launch_history": [
    {"start": "2026-05-20T04:12:47Z", "end": "2026-05-20T04:13:00Z"},
    {"start": "2026-05-20T04:14:00Z", "end": null}
  ],
  "deleted_at": null
}
```
