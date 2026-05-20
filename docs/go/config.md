# Go Package: `internal/config/paths.go`

## Purpose

Provides platform-aware directory path resolution for all mongoman operations. This is the foundation package — every other package depends on it.

## Algorithm

```
DefaultPaths()
1. Get $HOME via os.UserHomeDir()
2. Build ~/mongoman as base directory
3. Determine config directory:
   - Windows: %APPDATA%\mongoman (fallback: ~/AppData/Roaming/mongoman)
   - Unix:    $XDG_CONFIG_HOME/mongoman (fallback: ~/.config/mongoman)
4. Create all directories (data, logs, backups, config)
5. Return Paths struct
```

## Paths Struct

```go
type Paths struct {
    DataDir    string
    LogsDir    string
    BackupsDir string
    ConfigDir  string
}
```

## Helper Methods

| Method | Returns | Description |
|--------|---------|-------------|
| `MetaFile(name)` | `{ConfigDir}/{name}.json` | Instance metadata file |
| `DBPath(name)` | `{DataDir}/{name}` | Instance data directory |
| `LogPath(name)` | `{LogsDir}/{name}.log` | Instance log file |
| `BackupPath(name, ts)` | `{BackupsDir}/{name}_{ts}.tar.gz` or `.zip` | Backup archive |

## Platform Detection

Uses `runtime.GOOS` to determine Windows vs Unix paths. Backup extension is `.tar.gz` on Unix, `.zip` on Windows.
