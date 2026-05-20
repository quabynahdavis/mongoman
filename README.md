# 🗄️ mongoman — Cross-Platform MongoDB Instance Manager

A modular MongoDB instance manager for **Linux, macOS, BSD, and Windows**. Add, launch, clone, rename, reconfigure, backup, and control multiple MongoDB instances with ease—whether using direct process control or OS-native services (systemd, launchd, Windows Service).

## ✨ Features

- **Add** and **launch** MongoDB instances with custom ports
- **Rename**, **reconfigure**, **clone**, and **delete** instances
- **OS-native services**: systemd (Linux), launchd (macOS), Windows Service, with BSD fallback
- **Start/stop/restart** via systemd, launchd, or Windows Service control
- **View status**, **logs**, and **metadata** for each instance
- **Compressed backups** of data and metadata (tar.gz / zip)
- **mongosh shell** integration directly from the CLI
- **Launch history** tracking with JSON metadata
- **Cross-platform**: single binary for every OS

## 🚀 Quick Start

### Installation

```bash
# Option 1: Build and install from source (requires Go 1.22+)
./install.sh

# Option 2: Build with Make
make build && sudo make install

# Option 3: Go install directly
go install github.com/davisdeveloper/mongoman@latest
```

**Windows (PowerShell):**
```powershell
.\install.ps1
```

### Usage

```bash
# Add an instance
mongoman add dev27018 27018

# Launch it directly (forked process)
mongoman launch dev27018

# Enable as OS service
mongoman enable dev27018

# View status
mongoman status

# Tail logs
mongoman logs dev27018

# Kill direct process
mongoman kill dev27018

# Delete instance
mongoman delete dev27018
```

## 📋 Commands

| Command | Description |
|---------|-------------|
| `add <name> <port>` | Add new instance |
| `launch <name>` | Launch instance directly (forked) |
| `kill <name>` | Kill direct process |
| `delete <name>` | Delete instance and service |
| `rename <old> <new>` | Rename instance and service |
| `reconfigure <name> <port>` | Change port and update service |
| `clone <old> <new> <port>` | Clone instance to new name/port |
| `backup <name>` | Create compressed backup |
| `list` | List all instances |
| `status` | Show running/enabled/dead status |
| `logs <name>` | Tail log file live |
| `info <name>` | Show metadata and service status |
| `history <name>` | Show launch history |
| `shell <name>` | Launch mongosh for instance |
| `enable <name>` | Enable as OS service (systemd/launchd/Windows) |
| `disable <name>` | Disable OS service |
| `start <name>` | Start OS service |
| `stop <name>` | Stop OS service |
| `restart <name>` | Restart OS service |
| `help` | Show this help message |

## 📁 Directory Layout

| Content | Unix (Linux/macOS/BSD) | Windows |
|---------|------------------------|---------|
| Data | `~/mongoman/data/<name>` | `%USERPROFILE%\mongoman\data\<name>` |
| Logs | `~/mongoman/logs/<name>.log` | `%USERPROFILE%\mongoman\logs\<name>.log` |
| Backups | `~/mongoman/backups/` | `%USERPROFILE%\mongoman\backups\` |
| Config/Meta | `~/.config/mongoman/<name>.json` | `%APPDATA%\mongoman\<name>.json` |

## 🛠 Requirements

- **MongoDB** installed (`mongod` in PATH)
- **Go 1.22+** (only for building from source)
- **Platform-specific**:
  - Linux: systemd (optional, for service management)
  - macOS: launchd (built-in)
  - Windows: PowerShell 5+ (for installer), admin privileges for service control
  - BSD: direct process control (service management via stub)

## 🔧 Building from Source

```bash
# Build for current platform
make build

# Cross-compile for all supported platforms
make cross

# Create release archives
make release
```

### Supported Platforms

| OS | Architectures |
|----|---------------|
| Linux | amd64, arm64, 386 |
| macOS | amd64, arm64 (Apple Silicon) |
| Windows | amd64, 386 |
| FreeBSD | amd64 |
| OpenBSD | amd64 |
| NetBSD | amd64 |

## 📦 Repository Installation (Future)

Since mongoman is a single Go binary, it can be distributed via:

- **Homebrew** (macOS/Linux): `brew install mongoman`
- **Scoop** (Windows): `scoop install mongoman`
- **Snap** (Linux): `snap install mongoman`
- **Direct download**: Pre-built binaries from GitHub Releases
- **Go install**: `go install github.com/davisdeveloper/mongoman@latest`

## 🧑‍💻 Author

Built by [Davisville](https://github.com/davisdeveloper) — for developers who demand modularity, control, and elegance.

## 📄 License

MIT
