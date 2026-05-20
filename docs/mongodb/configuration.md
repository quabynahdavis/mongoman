# MongoDB Configuration

## Overview

mongoman manages multiple MongoDB instances by launching `mongod` with specific arguments for each instance. It does not modify MongoDB's system configuration file (`/etc/mongod.conf`).

## mongod Arguments

Each instance is launched with these arguments:

| Argument | Value | Description |
|----------|-------|-------------|
| `--port` | Instance port | Unique port per instance (e.g., 27018, 27019) |
| `--dbpath` | `~/mongoman/data/<name>` | Instance-specific data directory |
| `--bind_ip` | `localhost` | Restrict to local connections only |
| `--logpath` | `~/mongoman/logs/<name>.log` | Instance-specific log file |
| `--fork` | (flag) | Daemonize process (Unix only) |

## Requirements

- MongoDB must be installed (`mongod` must be in PATH)
- Tested with MongoDB 4.4+ (should work with any version supporting these flags)
- Each instance needs a unique port (default MongoDB port is 27017)

## mongosh Integration

`mongoman shell <name>` launches `mongosh` connected to the instance:

```
mongosh --port <instance-port>
```

This requires:
- `mongosh` in PATH (MongoDB Shell 1.0+)
- Instance must be running (checked via `proc.IsRunning()`)

## Data Directory Structure

```
~/mongoman/data/
├── dev27018/
│   ├── collection-0-<id>.wt
│   ├── collection-1-<id>.wt
│   ├── index-0-<id>.wt
│   ├── journal/
│   ├── _mdb_catalog.wt
│   ├── mongod.lock
│   ├── sizeStorer.wt
│   ├── storage.bson
│   └── WiredTiger*
└── test27019/
    └── ...
```

Each instance gets its own complete MongoDB data directory with WiredTiger storage engine.
