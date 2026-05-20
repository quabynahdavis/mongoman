# Process Management — Launch & Kill

## Launch Algorithm

When `mongoman launch <name>` is called:

```
1. Load instance metadata from JSON
2. Build mongod arguments:
   --port <port> --dbpath <dataDir> --bind_ip localhost --logpath <logPath> --fork
3. Execute mongod as a subprocess
4. If successful: record launch in metadata (timestamp + counter)
5. If failed: return error with mongod output
```

### mongod Arguments

| Argument | Value | Purpose |
|----------|-------|---------|
| `--port` | From metadata | Instance port |
| `--dbpath` | `~/mongoman/data/<name>` | Data directory |
| `--bind_ip` | `localhost` | Security — only local connections |
| `--logpath` | `~/mongoman/logs/<name>.log` | Log destination |
| `--fork` | (flag) | Daemonize (Unix only) |

## Kill Algorithm

When `mongoman kill <name>` is called:

```
1. Load instance metadata
2. Get port from metadata
3. Find PID: pgrep -f "mongod.*--port <port>"
4. Parse first result as integer PID
5. os.FindProcess(pid) — verify process exists
6. proc.Kill() — send SIGKILL
7. Record kill in metadata (end timestamp on latest launch)
8. Print "Killed MongoDB process for <name> (PID <pid>)"
```

## Status Detection

Two methods are combined for status:
1. **Direct process**: `proc.IsRunning()` — checks if mongod is running on the instance port
2. **Service status**: `service.Status()` — checks systemd/launchd/Windows Service status

In `cmdStatus()` and `cmdInfo()`:
```
if service says Stopped but proc says Running -> show Running
```
