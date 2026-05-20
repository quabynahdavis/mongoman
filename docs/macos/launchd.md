# macOS Service Management — launchd

## Overview

On macOS, mongoman uses launchd for OS-native service management. The `launchdManager` struct in `service_darwin.go` handles creation, control, and removal of launchd plist files.

## Plist File Generation

When `mongoman enable <name>` is called, a launchd plist is generated:

```
Path: ~/Library/LaunchAgents/com.mongoman.<name>.plist
```

### Plist Template

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.mongoman.<name></string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/mongod</string>
        <string>--port</string>
        <string><port></string>
        <string>--dbpath</string>
        <string><dbPath></string>
        <string>--bind_ip</string>
        <string>localhost</string>
        <string>--logpath</string>
        <string><logPath></string>
        <string>--fork</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string><logPath></string>
    <key>StandardErrorPath</key>
    <string><logPath></string>
</dict>
</plist>
```

**Key details:**
- Label follows reverse-DNS convention: `com.mongoman.<name>`
- `RunAtLoad` — starts on user login
- `KeepAlive` — automatically restarts if process crashes
- Installed in user's `LaunchAgents` (no root required)

## launchctl Commands

| Action | launchctl Command |
|--------|-------------------|
| Enable | `launchctl load ~/Library/LaunchAgents/com.mongoman.<name>.plist` |
| Disable | `launchctl unload <plist> && rm <plist>` |
| Start | `launchctl load <plist>` |
| Stop | `launchctl unload <plist>` |
| Restart | `launchctl unload <plist> && launchctl load <plist>` |

## Status Detection

```
1. launchctl list com.mongoman.<name> -> parse PID from output
   - If PID is numeric (not "-"): StatusRunning
2. Check if plist file exists on disk  -> StatusEnabled
3. Otherwise: StatusStopped
```
