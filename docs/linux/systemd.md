# Linux Service Management — systemd

## Overview

On Linux, mongoman uses systemd for OS-native service management. The `systemdManager` struct in `service_linux.go` handles creation, control, and removal of systemd service units.

## Service Unit Generation

When `mongoman enable <name>` is called, a systemd unit file is generated:

```
Path: /etc/systemd/system/mongod-<name>.service
```

### Unit File Template

```ini
[Unit]
Description=MongoDB Instance - <name>
After=network.target

[Service]
Type=forking
ExecStart=/usr/bin/mongod --port <port> --dbpath <dbPath> \
    --bind_ip localhost --logpath <logPath> --fork
ExecStop=/usr/bin/mongod --shutdown --dbpath <dbPath>
PIDFile=<dbPath>/mongod.pid
User=<current_user>
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

**Key details:**
- `Type=forking` — matches mongod's `--fork` behavior
- `ExecStop` uses mongod's built-in `--shutdown` for clean termination
- `User` is set to the current user (not root)
- `Restart=on-failure` for resilience

## systemd Commands

| Action | systemctl Command |
|--------|-------------------|
| Enable | `systemctl daemon-reload && systemctl enable <svc> && systemctl start <svc>` |
| Disable | `systemctl stop <svc> && systemctl disable <svc> && rm unit file && systemctl daemon-reload` |
| Start | `systemctl start <svc>` |
| Stop | `systemctl stop <svc>` |
| Restart | `systemctl restart <svc>` |

## Status Detection

```
1. systemctl is-active --quiet <svc>  -> if success: StatusRunning
2. systemctl is-enabled --quiet <svc> -> if success: StatusEnabled
3. Check if unit file exists on disk   -> if yes: StatusDisabled
4. Otherwise: StatusStopped
```

## Root Privileges

Writing to `/etc/systemd/system/` requires root. Users must run `enable`/`disable` with `sudo`.
