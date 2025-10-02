
---

## 📦 mongoInstance

A modular MongoDB instance manager for Linux and Windows. Add, launch, clone, rename, reconfigure, backup, and control multiple MongoDB instances with ease—whether you're using direct process control or systemd services.

---

### 🧠 Features

- Add and launch MongoDB instances with custom ports
- Rename, reconfigure, clone, and delete instances
- Enable systemd services for persistent startup (Linux)
- Start/stop/restart via systemd or direct process control
- View status, logs, and metadata for each instance
- Create compressed backups of data and metadata
- Cross-platform: Bash (Linux) + PowerShell (Windows)

---
## ✅ Installation

```bash
# Linux install
sudo chmod +x ./linux.sh
./linux.sh
```

```powershell
# Windows usage
# Note: Windows version hasn't been tested (Since I'm a full-time Arch Linux user). Contact me for if any issue or to suggest a fix
Set-Alias mongoInstance windows\mongoInstance.ps1
```
---

### 🚀 Quick Start (Linux)

```bash
# Add an instance
mongoInstance -add dev27018 27018

# Launch it directly (forked process)
mongoInstance dev27018

# Enable as systemd service
mongoInstance -setDefault dev27018

# View status
mongoInstance -status

# Tail logs
mongoInstance -logs dev27018

# Kill direct process
mongoInstance -kill dev27018

# Delete instance
mongoInstance -delete dev27018
```

---

### 🖥️ Quick Start (Windows PowerShell)

```powershell
# Add an instance
.\mongoInstance.ps1 -add dev27018 27018

# Launch it directly
.\mongoInstance.ps1 dev27018

# Enable as Windows service
.\mongoInstance.ps1 -setDefault dev27018

# View status
.\mongoInstance.ps1 -status

# Tail logs
.\mongoInstance.ps1 -logs dev27018

# Kill direct process
.\mongoInstance.ps1 -kill dev27018

# Delete instance
.\mongoInstance.ps1 -delete dev27018
```

---

### 🧪 Supported Commands

| Command                  | Description                                      |
|--------------------------|--------------------------------------------------|
| `-add name port`         | Add new instance                                |
| `name`                   | Launch instance directly                        |
| `-delete name`           | Delete instance and service                     |
| `-setDefault name`       | Enable as systemd/Windows service               |
| `-start/-stop/-restart`  | Control systemd/Windows service                 |
| `-kill name`             | Kill direct process                             |
| `-rename old new`        | Rename instance and service                     |
| `-reconfigure name port` | Change port and update service                  |
| `-clone old new port`    | Clone instance to new name/port                 |
| `-backup name`           | Create tar.gz or zip backup                     |
| `-list`                  | List all instances                              |
| `-status`                | Show running/enabled/dead status                |
| `-logs name`             | Tail log file live                              |
| `-info name`             | Show metadata and service status                |

---

### 📁 File Structure

- Linux metadata: `~/.mongo-meta/instanceName.conf`
- Linux data: `/data/instanceName`
- Linux logs: `/data/instanceName/mongod.log`
- Windows metadata: `%USERPROFILE%\.mongo-meta\instanceName.conf`
- Windows data: `C:\MongoData\instanceName`
- Windows logs: `C:\MongoData\instanceName\mongod.log`
- Backups: `~/mongo-backups` or `%USERPROFILE%\mongo-backups`

---

### 🛠 Requirements

#### Linux
- MongoDB installed (`mongod` in PATH)
- Systemd enabled
- Bash 4+

#### Windows
- MongoDB installed (`mongod.exe` in PATH)
- PowerShell 5+ or PowerShell Core
- Admin privileges for service control

---

### 🧑‍💻 Author

Built by [Davisville](https://github.com/davisdeveloper) — for developers who demand modularity, control, and elegance.

---