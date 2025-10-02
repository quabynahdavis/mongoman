#!/usr/bin/env pwsh

param(
    [string]$Command,
    [string]$Arg1,
    [string]$Arg2,
    [string]$Arg3,
    [string]$Arg4
)

Set-StrictMode -Version 3.0
$ErrorActionPreference = "Stop"

# === CONFIGURATION ===
$META_DIR = "$HOME\.mongo-meta"           # Stores instance metadata (port info)
$DB_ROOT = "$HOME\data"                   # Root directory for MongoDB data
$BACKUP_DIR = "$HOME\mongo-backups"       # Where backups are stored

# Ensure required directories exist
New-Item -ItemType Directory -Force -Path $META_DIR, $BACKUP_DIR, $DB_ROOT | Out-Null

# === HELPERS ===
function meta_file($name) { return "$META_DIR\$name.conf" }
function db_path($name) { return "$DB_ROOT\$name" }
function log_path($name) { return "$(db_path $name)\mongod.log" }
function service_name($name) { return "MongoDB-$name" }

# Reads the stored port from metadata
function read_port($name) {
    $mf = meta_file $name
    if (Test-Path $mf) {
        $content = Get-Content $mf -Raw
        if ($content -match "PORT=(\d+)") {
            return $matches[1]
        }
    }
    return $null
}

# Ensures instance metadata exists
function require_instance_exists($name) {
    if (-not (Test-Path (meta_file $name))) {
        Write-Error "Error: instance '$name' not found."
        exit 1
    }
}

# Check if MongoDB process is running on specified port
function Test-MongoDBRunning($port) {
    try {
        $process = Get-Process mongod -ErrorAction SilentlyContinue | Where-Object {
            $_.CommandLine -match "--port\s+$port"
        }
        return ($process -ne $null)
    }
    catch {
        return $false
    }
}

# Get MongoDB process by port
function Get-MongoDBProcess($port) {
    return Get-Process mongod -ErrorAction SilentlyContinue | Where-Object {
        $_.CommandLine -match "--port\s+$port"
    }
}

# === COMMANDS ===

# Adds a new instance with given name and port
function cmd_add {
    param($name, $port)
    
    if (-not $port) {
        Write-Error "Usage: mongoInstance -add instanceName portNumber"
        exit 1
    }
    
    "PORT=$port" | Out-File -FilePath (meta_file $name) -Encoding ASCII
    New-Item -ItemType Directory -Force -Path (db_path $name) | Out-Null
    New-Item -ItemType File -Force -Path (log_path $name) | Out-Null
    Write-Host "Added instance '$name' at $(db_path $name) (port $port)"
}

# Launches an instance using stored port
function cmd_launch {
    param($name)
    
    require_instance_exists $name
    $port = read_port $name
    $dbPath = db_path $name
    $logPath = log_path $name
    
    $arguments = @(
        "--port", $port
        "--dbpath", "`"$dbPath`""
        "--bind_ip", "localhost"
        "--fork"
        "--logpath", "`"$logPath`""
    )
    
    Start-Process -FilePath "mongod" -ArgumentList $arguments -NoNewWindow -Wait
    Write-Host "Launched '$name' on port $port"
}

# Stops the running process for instance
function cmd_kill {
    param($name)
    
    require_instance_exists $name
    $port = read_port $name
    
    $process = Get-MongoDBProcess $port
    if ($process) {
        $process | Stop-Process -Force
        Write-Host "Killed MongoDB process for '$name' (PID $($process.Id))"
    } else {
        Write-Host "No running process found for '$name' on port $port"
    }
}

# Starts a Windows service for the instance if enabled
function cmd_start {
    param($name)
    
    require_instance_exists $name
    $svc = service_name $name
    
    if (Get-Service -Name $svc -ErrorAction SilentlyContinue) {
        Start-Service -Name $svc
        Write-Host "Started MongoDB instance '$name'"
    } else {
        Write-Host "Instance '$name' is not enabled as a Windows service."
        Write-Host "Run 'mongoInstance setDefault $name' and try again"
    }
}

# Stops the Windows service for the instance
function cmd_stop {
    param($name)
    
    require_instance_exists $name
    $svc = service_name $name
    
    if (Get-Service -Name $svc -ErrorAction SilentlyContinue) {
        Stop-Service -Name $svc
        Write-Host "Stopped MongoDB instance '$name'"
    } else {
        Write-Host "Instance '$name' is not enabled as a Windows service."
        Write-Host "Run 'mongoInstance setDefault $name' and try again"
    }
}

# Restart the Windows service for the instance
function cmd_restart {
    param($name)
    
    require_instance_exists $name
    $svc = service_name $name
    
    if (Get-Service -Name $svc -ErrorAction SilentlyContinue) {
        Restart-Service -Name $svc
        Write-Host "Restarted MongoDB instance '$name'"
    } else {
        Write-Host "Instance '$name' is not enabled as a Windows service."
    }
}

# Deletes an instance and its Windows service
function cmd_delete {
    param($name)
    
    require_instance_exists $name
    $svc = service_name $name
    
    # Stop and remove service if it exists
    if (Get-Service -Name $svc -ErrorAction SilentlyContinue) {
        Stop-Service -Name $svc -Force -ErrorAction SilentlyContinue
        sc.exe delete $svc
    }
    
    # Remove files and directories
    Remove-Item -Recurse -Force (db_path $name) -ErrorAction SilentlyContinue
    Remove-Item -Force (meta_file $name) -ErrorAction SilentlyContinue
    
    Write-Host "Deleted instance '$name'"
}

# Enables instance as a Windows service
function cmd_setDefault {
    param($name)
    
    require_instance_exists $name
    $port = read_port $name
    $dbp = db_path $name
    $logp = log_path $name
    $svc = service_name $name
    
    $serviceArgs = @(
        "--port", $port
        "--dbpath", "`"$dbp`""
        "--bind_ip", "localhost"
        "--logpath", "`"$logp`""
    ) -join " "
    
    # Create Windows service
    New-Service -Name $svc `
                -BinaryPathName "`"mongod`" $serviceArgs" `
                -DisplayName "MongoDB Instance $name" `
                -StartupType Automatic `
                -ErrorAction Stop
    
    Start-Service -Name $svc
    Write-Host "Enabled and started Windows service $svc (port $port)"
}

# Lists all tracked instances and their ports
function cmd_list {
    Write-Host "📦 MongoDB Instances:"
    Get-ChildItem "$META_DIR\*.conf" | ForEach-Object {
        $name = $_.BaseName
        $port = read_port $name
        Write-Host " - $name (port: $port)"
    }
}

# Shows status of each instance (running, enabled, dead)
function cmd_status {
    Write-Host "📊 MongoDB Instance Status:"
    Get-ChildItem "$META_DIR\*.conf" | ForEach-Object {
        $name = $_.BaseName
        $port = read_port $name
        $svc = service_name $name
        
        $status = "❌ Dead"
        if (Test-MongoDBRunning $port) {
            $status = "✅ Running"
        } elseif (Get-Service -Name $svc -ErrorAction SilentlyContinue) {
            $status = "🔁 Enabled"
        }
        
        Write-Host " - $name (port: $port): $status"
    }
}

# Renames an instance and its metadata/service
function cmd_rename {
    param($old, $new)
    
    require_instance_exists $old
    
    # Rename metadata file
    Rename-Item (meta_file $old) (meta_file $new)
    
    # Rename data directory
    if (Test-Path (db_path $old)) {
        Rename-Item (db_path $old) (db_path $new)
    }
    
    # Handle service renaming
    $oldSvc = service_name $old
    $newSvc = service_name $new
    
    if (Get-Service -Name $oldSvc -ErrorAction SilentlyContinue) {
        Stop-Service -Name $oldSvc -Force -ErrorAction SilentlyContinue
        sc.exe delete $oldSvc
        
        # Recreate service with new name
        cmd_setDefault $new
    }
    
    Write-Host "Renamed '$old' to '$new'"
}

# Updates the port of an instance and rewrites service if needed
function cmd_reconfigure {
    param($name, $newport)
    
    require_instance_exists $name
    "PORT=$newport" | Out-File -FilePath (meta_file $name) -Encoding ASCII
    
    $svc = service_name $name
    if (Get-Service -Name $svc -ErrorAction SilentlyContinue) {
        cmd_setDefault $name
    }
    
    Write-Host "Reconfigured '$name' to port $newport"
}

# Clones an instance to a new name and port
function cmd_clone {
    param($old, $new, $newport)
    
    require_instance_exists $old
    
    # Copy data directory
    Copy-Item -Recurse -Path (db_path $old) -Destination (db_path $new)
    
    # Create new metadata
    "PORT=$newport" | Out-File -FilePath (meta_file $new) -Encoding ASCII
    
    # Create service if original had one
    if (Get-Service -Name (service_name $old) -ErrorAction SilentlyContinue) {
        cmd_setDefault $new
    }
    
    Write-Host "Cloned '$old' to '$new' (port $newport)"
}

# Creates a zip backup of the instance
function cmd_backup {
    param($name)
    
    require_instance_exists $name
    $timestamp = Get-Date -Format "yyyyMMddHHmmss"
    $outFile = "$BACKUP_DIR\${name}_${timestamp}.zip"
    
    $files = @(
        (db_path $name),
        (meta_file $name)
    )
    
    Compress-Archive -Path $files -DestinationPath $outFile -CompressionLevel Optimal
    Write-Host "Backup created: $outFile"
}

# Logs the activities of the instance
function cmd_logs {
    param($name)
    
    require_instance_exists $name
    $log = log_path $name
    
    if (Test-Path $log) {
        Write-Host "Tailing logs for '$name' (Ctrl+C to exit):"
        Get-Content $log -Wait
    } else {
        Write-Host "Log file not found for '$name'"
    }
}

# Displays the info of the instance
function cmd_info {
    param($name)
    
    require_instance_exists $name
    $port = read_port $name
    $dbp = db_path $name
    $logp = log_path $name
    $svc = service_name $name
    
    $status = "❌ Dead"
    if (Test-MongoDBRunning $port) {
        $status = "✅ Running"
    } elseif (Get-Service -Name $svc -ErrorAction SilentlyContinue) {
        $status = "🔁 Enabled"
    }
    
    Write-Host "📋 Instance Info: $name"
    Write-Host " - Port: $port"
    Write-Host " - Data Path: $dbp"
    Write-Host " - Log Path: $logp"
    Write-Host " - Windows Service: $svc"
    Write-Host " - Status: $status"
}

# Displays the help and usage menu
function cmd_help {
    Write-Host "Available options:"
    Write-Host "  add name port          - Add new instance"
    Write-Host "  launch name            - Launch instance directly"
    Write-Host "  delete name            - Delete instance and service"
    Write-Host "  setDefault name        - Enable as Windows service"
    Write-Host "  start/stop/restart name - Control Windows service"
    Write-Host "  kill name              - Kill direct process"
    Write-Host "  rename old new         - Rename instance and service"
    Write-Host "  reconfigure name port  - Change port and update service"
    Write-Host "  clone old new port     - Clone instance to new name/port"
    Write-Host "  backup name            - Create zip backup"
    Write-Host "  list                   - List all instances"
    Write-Host "  status                 - Show running/enabled/dead status"
    Write-Host "  logs name              - Tail log file live"
    Write-Host "  info name              - Show metadata and service status"
}

# === DISPATCHER ===
try {
    switch ($Command) {
        "add"         { cmd_add $Arg1 $Arg2 }
        "delete"      { cmd_delete $Arg1 }
        "setDefault"  { cmd_setDefault $Arg1 }
        "list"        { cmd_list }
        "status"      { cmd_status }
        "rename"      { cmd_rename $Arg1 $Arg2 }
        "reconfigure" { cmd_reconfigure $Arg1 $Arg2 }
        "clone"       { cmd_clone $Arg1 $Arg2 $Arg3 }
        "backup"      { cmd_backup $Arg1 }
        "start"       { cmd_start $Arg1 }
        "stop"        { cmd_stop $Arg1 }
        "restart"     { cmd_restart $Arg1 }
        "kill"        { cmd_kill $Arg1 }
        "launch"      { cmd_launch $Arg1 }
        "logs"        { cmd_logs $Arg1 }
        "info"        { cmd_info $Arg1 }
        ""            { 
            if ($Arg1) { 
                cmd_launch $Arg1 
            } else { 
                Write-Host "Usage: mongoInstance [options]"
                cmd_help
            }
        }
        default       { cmd_help }
    }
}
catch {
    Write-Error "Error: $($_.Exception.Message)"
    exit 1
}