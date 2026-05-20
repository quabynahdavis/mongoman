# Go Package: `internal/proc/proc.go`

## Purpose

Manages mongod process lifecycle — launching forked daemons and killing them by port.

## Functions

### Launch(inst)

Starts a mongod process as a forked daemon for the given instance.

**Algorithm:**
```
1. Build mongod argument list:
   --port <port> --dbpath <path> --bind_ip localhost --logpath <path> --fork
2. Execute mongod via exec.Command
3. If successful, record launch in instance metadata
4. Return error if mongod fails to start
```

**Platform Notes:**
- `--fork` flag tells mongod to daemonize on Unix
- On Windows, `--fork` is ignored by mongod (still works with Start-Process)
- mongod must be in PATH

### Kill(inst)

Terminates the mongod process associated with the instance's port.

**Algorithm:**
```
1. Build port string from instance metadata
2. Call findPIDByPort(port) to get the process PID
3. os.FindProcess(pid) — locate the process
4. proc.Kill() — terminate it
5. Record kill timestamp in instance metadata
6. Print killed process info to stdout
```

### IsRunning(inst) bool

Checks if a mongod process is running on the instance's port.

```
1. Call findPIDByPort(port)
2. Return true if PID found, false otherwise
```

### findPIDByPort(port) — Internal Helper

Uses `pgrep` to find a mongod process by its `--port` argument.

**Algorithm:**
```
1. Execute: pgrep -f "mongod.*--port <port>"
2. Parse first line of output as integer PID
3. Return PID or error "process not found"
```

**Platform Note:** `pgrep` is available on all Unix systems (Linux, macOS, BSD). On Windows, this would need a different implementation using `Get-Process`.
