# Launch History Tracking

## Overview

mongoman maintains a detailed launch history for each instance. This includes every time an instance was launched and killed, with timestamps.

## Data Structure

Each instance's JSON metadata file contains:

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

### LaunchRecord

```go
type LaunchRecord struct {
    Start time.Time  `json:"start"`
    End   *time.Time `json:"end,omitempty"`
}
```

- `Start` — set when `mongoman launch` is called (via `RecordLaunch()`)
- `End` — set when `mongoman kill` is called (via `RecordKill()`)
- `End` is `null` if the process is still running

## Recording Flow

### Launch
```
1. mongoman launch <name>
2. mongod starts successfully
3. instance.RecordLaunch() is called:
   a. Meta.LaunchCount++
   b. Append LaunchRecord{Start: time.Now()} to Meta.LaunchHistory
   c. Save metadata to disk
```

### Kill
```
1. mongoman kill <name>
2. mongod process terminated
3. instance.RecordKill() is called:
   a. Get the last LaunchRecord in Meta.LaunchHistory
   b. Set its End field to time.Now()
   c. Save metadata to disk
```

### Delete
```
1. mongoman delete <name>
2. instance.MarkDeleted() is called:
   a. Set Meta.DeletedAt to time.Now()
   b. Save metadata to disk
```

## Display

The `mongoman history <name>` command outputs the full metadata as formatted JSON:

```bash
$ mongoman history dev27018
{
  "name": "dev27018",
  "port": 27018,
  "created_at": "2026-05-20T04:12:47.258014518Z",
  "launch_count": 3,
  "launch_history": [...],
  "deleted_at": null
}
```

## Use Cases

- **Audit trail**: Know exactly when each instance was started/stopped
- **Uptime tracking**: Calculate duration from start-end pairs
- **Usage monitoring**: Launch count shows how often an instance is used
- **Debugging**: Identify if an instance was killed unexpectedly (missing `end` timestamp)
