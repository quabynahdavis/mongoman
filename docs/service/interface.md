# Service Manager Interface

## Overview

The service package provides a unified interface for OS-native service management across all supported platforms. Each platform implements this interface using its native service system.

## Interface Definition

```go
type Manager interface {
    Enable(name string, port int, dbPath, logPath string) error
    Disable(name string) error
    Start(name string) error
    Stop(name string) error
    Restart(name string) error
    Status(name string) Status
    IsEnabled(name string) bool
}
```

## Status Type

```go
type Status int

const (
    StatusUnknown Status = iota
    StatusRunning
    StatusStopped
    StatusEnabled
    StatusDisabled
)
```

The `Status` type has a `String()` method for display: "✅ Running", "❌ Stopped", "🔁 Enabled", "⛔ Disabled", "❓ Unknown".

## Factory Function

```go
func NewManager() Manager {
    return newPlatformManager()
}
```

The `newPlatformManager()` function is implemented in each platform-specific file using Go build tags.

## Build Tag Selection

| File | Build Tag | Platform | Implementation |
|------|-----------|----------|----------------|
| `service_linux.go` | `//go:build linux` | Linux | systemdManager |
| `service_darwin.go` | `//go:build darwin` | macOS | launchdManager |
| `service_windows.go` | `//go:build windows` | Windows | windowsManager |
| `service_stub.go` | `//go:build !linux && !darwin && !windows` | BSD/Other | stubManager |

## How Build Tags Work

When compiling, Go only includes files whose build tags match the target OS. This means:
- On Linux: only `service.go` + `service_linux.go` are compiled
- On macOS: only `service.go` + `service_darwin.go` are compiled
- On Windows: only `service.go` + `service_windows.go` are compiled
- On BSD: only `service.go` + `service_stub.go` are compiled

The `newPlatformManager()` function is defined exactly once per build, avoiding symbol conflicts.
