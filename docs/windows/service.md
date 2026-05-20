# Windows Service Management

## Overview

On Windows, mongoman uses the Windows Service Control Manager via `sc.exe` for OS-native service management. The `windowsManager` struct in `service_windows.go` handles creation, control, and removal of Windows services.

## Service Creation

When `mongoman enable <name>` is called, a Windows service is created:

```
Service Name: MongoDB-<name>
Display Name: MongoDB Instance <name>
```

### sc.exe Command

```
sc.exe create MongoDB-<name> \
    binPath= "mongod --port <port> --dbpath \"<dbPath>\" \
             --bind_ip localhost --logpath \"<logPath>\"" \
    start= auto \
    DisplayName= "MongoDB Instance <name>"
```

**Key details:**
- `start= auto` — service starts automatically on boot
- `binPath` includes the full mongod command line
- Admin privileges required

## sc.exe Commands

| Action | Command |
|--------|---------|
| Enable | `sc.exe create <svc> ... && sc.exe start <svc>` |
| Disable | `sc.exe stop <svc> && sc.exe delete <svc>` |
| Start | `sc.exe start <svc>` |
| Stop | `sc.exe stop <svc>` |
| Restart | `sc.exe stop <svc> && sc.exe start <svc>` |

## Status Detection

```
1. sc.exe query MongoDB-<name>
   - If output contains "RUNNING": StatusRunning
   - If output contains "STOPPED": StatusStopped
2. sc.exe qc MongoDB-<name>
   - If success (service config exists): StatusDisabled
3. Otherwise: StatusStopped
```
