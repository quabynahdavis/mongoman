# Installation Guide

## Prerequisites

- **Go 1.22+** (for building from source)
- **MongoDB** installed (`mongod` in PATH)
- **Platform-specific**:
  - Linux: systemd (optional, for service management)
  - macOS: launchd (built-in)
  - Windows: PowerShell 5+ (for installer), admin privileges for service control
  - BSD: direct process control only

## Installation Methods

### Method 1: Install Script (Linux/macOS)

```bash
chmod +x install.sh
./install.sh
```

This script:
1. Checks Go is installed (1.22+)
2. Builds the binary with `CGO_ENABLED=0 go build`
3. Installs to `/usr/local/bin/mongoman`
4. Verifies the installation

### Method 2: Install Script (Windows PowerShell)

```powershell
.\install.ps1
```

This script:
1. Checks Go is installed
2. Builds the binary
3. Installs to `%USERPROFILE%\.mongoman\bin\mongoman.exe`
4. Adds the directory to user PATH

### Method 3: Makefile

```bash
make build       # Build for current platform
sudo make install # Install to /usr/local/bin
```

### Method 4: Go Install

```bash
go install github.com/davisdeveloper/mongoman@latest
```

## Cross-Compilation

Build for all 10 supported platforms:

```bash
make cross
```

Output in `build/<os>/<arch>/mongoman[.exe]`.

Create release archives:

```bash
make release
```

## Verification

```bash
mongoman help
mongoman add test 27018
mongoman list
mongoman delete test
```

## Uninstall

```bash
sudo rm /usr/local/bin/mongoman      # Linux/macOS
rm %USERPROFILE%\.mongoman\bin\mongoman.exe  # Windows
```
