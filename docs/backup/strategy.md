# Backup Strategy

## Overview

When `mongoman backup <name>` is called, a compressed archive of the instance's data directory is created.

## Algorithm

```
1. Load instance metadata (ensure exists)
2. Generate timestamp: time.Now().Format("20060102150405")
3. Build backup path: ~/mongoman/backups/<name>_<timestamp>.tar.gz
4. Execute: tar -czf <backupPath> -C <dataDir> <name>
5. Print confirmation: "Backup created: <path>"
```

## Platform Differences

| Platform | Format | Command |
|----------|--------|---------|
| Unix (Linux/macOS/BSD) | `.tar.gz` | `tar -czf <path> -C <dir> <name>` |
| Windows | `.zip` | Would use `Compress-Archive` |

## Backup Storage

All backups are stored in `~/mongoman/backups/`. The filename encodes:
- Instance name
- Timestamp (YYYYMMDDHHMMSS format)

Example: `dev27018_20260520150405.tar.gz`

## Restore

Restore is not yet implemented as a command. To manually restore:
```bash
cd ~/mongoman/backups
tar -xzf dev27018_20260520150405.tar.gz -C ~/mongoman/data/
mongoman reconfigure dev27018 <original-port>
```

## Future Enhancements

- `mongoman restore <name> <backup-file>` command
- Compression level options
- Backup rotation/cleanup policies
- Remote backup destinations (S3, SFTP)
