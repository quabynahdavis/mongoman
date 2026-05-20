# mongoman Go Rewrite — Architecture Plan

**Date:** 2026-05-20T04:08:32Z
**Status:** ✅ Implemented

## Problem Analysis

The original mongoman codebase had:
1. **Two divergent implementations** (monolithic Bash + modular Bash)
2. **14 bugs** including syntax errors, missing functions, broken dispatchers
3. **No cross-platform support** (Bash is Unix-only, PowerShell is Windows-only)
4. **Installer referencing non-existent files**

## Solution: Go Rewrite

### Why Go?

| Factor | Go | Bash | Python | Rust |
|--------|----|------|--------|------|
| Single binary | ✅ | ❌ | ❌ | ✅ |
| Cross-platform | ✅ | ❌ | ✅ | ✅ |
| Zero dependencies | ✅ | N/A | ❌ | ✅ |
| Repository installable | ✅ | ❌ | ✅ | ✅ |
| Easy to learn/maintain | ✅ | ✅ | ✅ | ❌ |

### Architecture Decisions

| Decision | Rationale |
|----------|-----------|
| Pure Go stdlib | Zero external dependencies |
| `flag` package alternative | Manual arg dispatch for subcommands |
| Build tags for services | `//go:build linux/darwin/windows` |
| JSON metadata | Structured, human-readable, easy to debug |
| XDG Base Dir (Unix) | Follows platform conventions |
| AppData (Windows) | Follows Windows conventions |

### Directory Layout (from plan.txt)

```
~/mongoman/
├── data/       # Instance data directories
├── logs/       # Instance log files
├── backups/    # Backup archives
~/.config/mongoman/  # Instance metadata (JSON)
```

### Implementation Order

1. ✅ `internal/config/paths.go` — Foundation (no deps)
2. ✅ `internal/instance/instance.go` — Core (depends on config)
3. ✅ `internal/proc/proc.go` — Process mgmt (depends on instance)
4. ✅ `internal/service/*.go` — Service mgmt (standalone package)
5. ✅ `main.go` — CLI (depends on all)
6. ✅ `Makefile` — Build system
7. ✅ `install.sh` + `install.ps1` — Installers
8. ✅ `README.md` — Documentation

### Testing Strategy

1. ✅ Native build (`go build ./...`)
2. ✅ Cross-compile (make cross — 10 platforms)
3. ✅ CRUD operations (add, list, status, rename, reconfigure, clone, info, history, delete)
4. ✅ Edge cases (empty list, duplicate add prevention)
