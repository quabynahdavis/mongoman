# Go Package: `internal/instance/instance.go`

## Purpose

Manages MongoDB instance lifecycle — creation, loading, updating, deletion, and metadata tracking.

## Types

### Metadata Struct

```go
type Metadata struct {
    Name          string         `json:"name"`
    Port          int            `json:"port"`
    CreatedAt     time.Time      `json:"created_at"`
    LaunchCount   int            `json:"launch_count"`
    LaunchHistory []LaunchRecord `json:"launch_history,omitempty"`
    DeletedAt     *time.Time     `json:"deleted_at,omitempty"`
}
```

### Instance Struct

Wraps Metadata together with resolved filesystem paths for quick access.

```go
type Instance struct {
    Meta     Metadata
    Paths    *config.Paths
    MetaPath string
    DBPath   string
    LogPath  string
}
```

## Functions

### Create(paths, name, port) — Add Instance

```
1. Check metadata file doesn't already exist (conflict detection)
2. Create data directory (os.MkdirAll)
3. Create empty log file (os.WriteFile)
4. Initialize Metadata struct with name, port, creation timestamp
5. Write JSON metadata to disk
6. Return Instance pointer
```

### Load(paths, name) — Read Instance

```
1. Read metadata file from disk
2. Unmarshal JSON into Metadata struct
3. Resolve DBPath, LogPath, MetaPath
4. Return Instance pointer (or error if not found)
```

### Rename(paths, oldName, newName)

```
1. Load old instance (panic if not found)
2. Rename metadata file:  os.Rename(oldMeta, newMeta)
3. Rename data directory: os.Rename(oldDB, newDB)
4. Update Instance fields (Name, MetaPath, DBPath, LogPath)
5. Save updated metadata
```

### Reconfigure(paths, name, newPort)

```
1. Load instance
2. Update Meta.Port = newPort
3. Save metadata
```

### Delete(paths, name)

```
1. Load instance
2. os.RemoveAll(data directory)
3. os.Remove(metadata file)
```

### Clone(paths, srcName, dstName, port)

```
1. Validate source exists
2. Copy data directory (copyDir via cp/robocopy)
3. Create destination instance with new name/port
```

### ListAll(paths)

```
1. ReadDir on ConfigDir
2. Filter for *.json files
3. Strip .json extension to get instance names
4. Return sorted list
```

## Launch History Tracking

- `RecordLaunch()` — Appends a LaunchRecord with `Start: time.Now()` to history, increments LaunchCount
- `RecordKill()` — Sets `End: time.Now()` on the most recent LaunchRecord
- `MarkDeleted()` — Sets `DeletedAt: time.Now()` for audit trail

## Error Handling

All errors are wrapped with context using `fmt.Errorf("context: %w", err)`. Missing instances produce a clear `"instance %q not found"` message. The `RequireExists()` helper exits with code 1 on missing instances for CLI commands.
