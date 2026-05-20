# CLI Commands Reference

## Overview

mongoman uses a subcommand-based interface. All commands are dispatched from `main.go` via a `switch` statement.

## Command Table

| Command | Arguments | Description | Package |
|---------|-----------|-------------|---------|
| `add` | `<name> <port>` | Add new MongoDB instance | instance.Create |
| `launch` | `<name>` | Launch instance (forked process) | proc.Launch |
| `kill` | `<name>` | Kill direct process | proc.Kill |
| `delete` | `<name>` | Delete instance and data | instance.Delete |
| `rename` | `<old> <new>` | Rename instance | instance.Rename |
| `reconfigure` | `<name> <port>` | Change port | instance.Reconfigure |
| `clone` | `<src> <dst> <port>` | Clone instance | instance.Clone |
| `backup` | `<name>` | Create backup archive | tar/zip via exec |
| `list` | (none) | List all instances | instance.ListAll |
| `status` | (none) | Show running/enabled status | service.Status + proc.IsRunning |
| `logs` | `<name>` | Tail log file | tail -f via exec |
| `info` | `<name>` | Show metadata and status | instance.Load + service.Status |
| `history` | `<name>` | Show launch history | JSON output of metadata |
| `shell` | `<name>` | Launch mongosh | mongosh via exec |
| `enable` | `<name>` | Enable as OS service | service.Enable |
| `disable` | `<name>` | Disable OS service | service.Disable |
| `start` | `<name>` | Start OS service | service.Start |
| `stop` | `<name>` | Stop OS service | service.Stop |
| `restart` | `<name>` | Restart OS service | service.Restart |
| `help` | (none) | Show help | printUsage() |

## Backward Compatibility

If no recognized command is provided, the first argument is treated as an instance name and `launch` is attempted. This matches the original Bash script behavior where `mongoInstance dev27018` would launch it directly.

## Dispatch Algorithm

```
dispatch(cmd, args):
  switch cmd:
    "add"         -> cmdAdd(args)         # expects [name, port]
    "launch"      -> cmdLaunch(args)       # expects [name]
    "kill"        -> cmdKill(args)         # expects [name]
    "delete"      -> cmdDelete(args)       # expects [name]
    "rename"      -> cmdRename(args)       # expects [old, new]
    "reconfigure" -> cmdReconfigure(args)  # expects [name, port]
    "clone"       -> cmdClone(args)        # expects [src, dst, port]
    "backup"      -> cmdBackup(args)       # expects [name]
    "list"        -> cmdList()             # no args
    "status"      -> cmdStatus()           # no args
    "logs"        -> cmdLogs(args)         # expects [name]
    "info"        -> cmdInfo(args)         # expects [name]
    "history"     -> cmdHistory(args)      # expects [name]
    "shell"       -> cmdShell(args)        # expects [name]
    "enable"      -> cmdEnable(args)       # expects [name]
    "disable"     -> cmdDisable(args)      # expects [name]
    "start"       -> cmdStart(args)        # expects [name]
    "stop"        -> cmdStop(args)         # expects [name]
    "restart"     -> cmdRestart(args)      # expects [name]
    "help"/"-h"   -> printUsage()
    default       -> try cmdLaunch([cmd])  # backward compat
```

## Error Handling

All command handlers return `error`. The dispatch function prints errors to stderr and exits with code 1:

```go
if err := dispatch(cmd, args); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```
