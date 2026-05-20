#!/usr/bin/env pwsh
# mongoman — Windows PowerShell Installer
# Builds and installs mongoman, adding it to the user's PATH.

$AppName = "mongoman"

function Write-Info  { Write-Host "✅ $($args[0])" -ForegroundColor Green }
function Write-Warn  { Write-Host "⚠️  $($args[0])" -ForegroundColor Yellow }
function Write-Error { Write-Host "❌ $($args[0])" -ForegroundColor Red }

# ── Pre-flight checks ─────────────────────────────────────────────────────────

# Check Go is installed
$goCmd = Get-Command "go" -ErrorAction SilentlyContinue
if (-not $goCmd) {
    Write-Error "Go is not installed. Please install Go 1.22+ from https://go.dev/dl/"
    exit 1
}

$goVersion = go version
Write-Info "Go detected: $goVersion"

# ── Build ─────────────────────────────────────────────────────────────────────

Write-Host ""
Write-Host "📦 Building $AppName..."

$buildDir = Join-Path $env:TEMP "mongoman-build"
if (Test-Path $buildDir) { Remove-Item -Recurse -Force $buildDir }
New-Item -ItemType Directory -Force -Path $buildDir | Out-Null

$binaryPath = Join-Path $buildDir "$AppName.exe"

$env:CGO_ENABLED = "0"
go build -ldflags="-s -w" -o $binaryPath

if (-not (Test-Path $binaryPath)) {
    Write-Error "Build failed"
    exit 1
}

Write-Info "Build successful"

# ── Install ───────────────────────────────────────────────────────────────────

Write-Host ""
Write-Host "📋 Installing $AppName..."

# Install to a user-local bin directory
$installDir = Join-Path $env:USERPROFILE ".mongoman\bin"
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
}

$installedPath = Join-Path $installDir "$AppName.exe"
Copy-Item -Path $binaryPath -Destination $installedPath -Force

Write-Info "Installed to $installedPath"

# ── Add to PATH ───────────────────────────────────────────────────────────────

$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notlike "*$installDir*") {
    $newPath = "$installDir;$userPath"
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    # Also update current session
    $env:PATH = "$installDir;$env:PATH"
    Write-Info "Added $installDir to your PATH (user-level)"
} else {
    Write-Info "$installDir is already in your PATH"
}

# ── Cleanup ───────────────────────────────────────────────────────────────────

Remove-Item -Recurse -Force $buildDir -ErrorAction SilentlyContinue

# ── Verify ────────────────────────────────────────────────────────────────────

Write-Host ""
$cmdCheck = Get-Command $AppName -ErrorAction SilentlyContinue
if ($cmdCheck) {
    Write-Info "$AppName is now available. Run '$AppName help' to get started."
    & $AppName help
} else {
    Write-Warn "Installation complete, but $AppName may not be in your current session's PATH."
    Write-Warn "Open a new terminal or run: `$env:PATH = `"$installDir;`$env:PATH`""
}

Write-Host ""
Write-Host "🎉 Installation complete!" -ForegroundColor Green
