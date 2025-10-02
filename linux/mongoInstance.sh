#!/bin/bash
set -euo pipefail

# === CONFIGURATION ===
META_DIR="$HOME/.mongo-meta"              # Stores instance metadata (port info)
DB_ROOT="/data"                           # Root directory for MongoDB data
BACKUP_DIR="$HOME/mongo-backups"         # Where backups are stored
mkdir -p "$META_DIR" "$BACKUP_DIR"       # Ensure required directories exist

# === HELPERS ===
meta_file() { echo "$META_DIR/$1.conf"; }
db_path() { echo "$DB_ROOT/$1"; }
log_path() { echo "$(db_path "$1")/mongod.log"; }
service_name() { echo "mongod-$1"; }
service_file() { echo "/etc/systemd/system/$(service_name "$1").service"; }

# Reads the stored port from metadata
read_port() {
    local mf="$(meta_file "$1")"
    if [ -f "$mf" ]; then
        source "$mf"
        echo "${PORT:-}"
    else
        echo ""
    fi
}

# Ensures instance metadata exists
require_instance_exists() {
    if [ ! -f "$(meta_file "$1")" ]; then
        echo "Error: instance '$1' not found." >&2
        exit 1
    fi
}

# === COMMANDS ===

# Adds a new instance with given name and port
cmd_add() {
    local name="$1" port="$2"
    if [ -z "$port" ]; then
        echo "Usage: mongoInstance -add instanceName portNumber" >&2
        exit 1
    fi
    echo "PORT=$port" > "$(meta_file "$name")"
    mkdir -p "$(db_path "$name")"
    touch "$(log_path "$name")"
    echo "Added instance '$name' at $(db_path "$name") (port $port)"
}

# Launches an instance using stored port
cmd_launch() {
    local name="$1"
    require_instance_exists "$name"
    local port="$(read_port "$name")"
    mongod --port "$port" --dbpath "$(db_path "$name")" --bind_ip localhost --fork --logpath "$(log_path "$name")"
    echo "Launched '$name' on port $port"
}

# Stops the running process for instance
cmd_kill() {
    local name="$1"
    require_instance_exists "$name"
    local port="$(read_port "$name")"
    local pid
    pid=$(pgrep -f "mongod.*--port $port")
    if [ -n "$pid" ]; then
        kill "$pid"
        echo "Killed MongoDB process for '$name' (PID $pid)"
    else
        echo "No running process found for '$name' on port $port"
    fi
}


# Starts a systemd service for the instance if enabled
cmd_start() {
    local name="$1"
    require_instance_exists "$name"
    local svc="$(service_name "$name")"
    if systemctl list-units --type=service | grep -q "$svc"; then
        sudo systemctl start "$svc"
        echo "Started MongoDB instance '$name'"
    else
        echo "Instance '$name' is not enabled as a systemd service."
        echo "Run 'mongoInstance -setDefault $name' and try again"
    fi
}

# Stops the systemd service for the instance started
cmd_stop() {
    local name="$1"
    require_instance_exists "$name"
    local svc="$(service_name "$name")"
    if systemctl list-units --type=service | grep -q "$svc"; then
        sudo systemctl stop "$svc"
        echo "Stopped MongoDB instance '$name'"
    else
        echo "Instance '$name' is not enabled as a systemd service."
        echo "Run 'mongoInstance -setDefault $name' and try again"
    fi
}

# Restart the systemd service for the instance
cmd_restart() {
    local name="$1"
    require_instance_exists "$name"
    local svc="$(service_name "$name")"
    if systemctl list-units --type=service | grep -q "$svc"; then
        sudo systemctl restart "$svc"
        echo "Restarted MongoDB instance '$name'"
    else
        echo "Instance '$name' is not enabled as a systemd service."
    fi
}


# Deletes an instance and its systemd service
cmd_delete() {
    local name="$1"
    require_instance_exists "$name"
    local svc="$(service_name "$name")"
    sudo systemctl stop "$svc" || true
    sudo systemctl disable "$svc" || true
    sudo rm -f "$(service_file "$name")"
    sudo systemctl daemon-reload
    rm -rf "$(db_path "$name")"
    rm -f "$(meta_file "$name")"
    echo "Deleted instance '$name'"
}

# Enables instance as a systemd service
cmd_setDefault() {
    local name="$1"
    require_instance_exists "$name"
    local port="$(read_port "$name")"
    local dbp="$(db_path "$name")"
    local logp="$(log_path "$name")"
    local svc="$(service_name "$name")"
    
  sudo tee "$(service_file "$name")" > /dev/null <<EOF
[Unit]
Description=MongoDB Instance $name
After=network.target

[Service]
ExecStart=/usr/bin/mongod --port $port --dbpath $dbp --bind_ip localhost --logpath $logp
User=mongodb
Group=mongodb
Restart=always
LimitNOFILE=64000

[Install]
WantedBy=multi-user.target
EOF
    
    sudo systemctl daemon-reload
    sudo systemctl enable "$svc"
    sudo systemctl restart "$svc"
    echo "Enabled and started systemd service $svc (port $port)"
}

# Lists all tracked instances and their ports
cmd_list() {
    echo "📦 MongoDB Instances:"
    for f in "$META_DIR"/*.conf; do
        [ -e "$f" ] || continue
        local name="$(basename "$f" .conf)"
        source "$f"
        echo " - $name (port: ${PORT:-unknown})"
    done
}

# Shows status of each instance (running, enabled, dead)
cmd_status() {
    echo "📊 MongoDB Instance Status:"
    for f in "$META_DIR"/*.conf; do
        [ -e "$f" ] || continue
        local name="$(basename "$f" .conf)"
        source "$f"
        local port="${PORT:-}"
        local svc="$(service_name "$name")"
        local status="❌ Dead"
        if pgrep -f "mongod.*--port $port" > /dev/null; then
            status="✅ Running"
            elif systemctl is-enabled "$svc" &> /dev/null; then
            status="🔁 Enabled"
        fi
        echo " - $name (port: $port): $status"
    done
}

# Renames an instance and its metadata/service
cmd_rename() {
    local old="$1" new="$2"
    require_instance_exists "$old"
    mv "$(meta_file "$old")" "$(meta_file "$new")"
    mv "$(db_path "$old")" "$(db_path "$new")"
    local oldsvc="$(service_name "$old")"
    local newsvc="$(service_name "$new")"
    if systemctl list-units --type=service | grep -q "$oldsvc"; then
        sudo systemctl stop "$oldsvc"
        sudo systemctl disable "$oldsvc"
        sudo mv "$(service_file "$old")" "$(service_file "$new")"
        sudo systemctl daemon-reload
        sudo systemctl enable "$newsvc"
        sudo systemctl start "$newsvc"
    fi
    echo "Renamed '$old' to '$new'"
}

# Updates the port of an instance and rewrites systemd if needed
cmd_reconfigure() {
    local name="$1" newport="$2"
    require_instance_exists "$name"
    echo "PORT=$newport" > "$(meta_file "$name")"
    local svc="$(service_name "$name")"
    if systemctl list-units --type=service | grep -q "$svc"; then
        cmd_setDefault "$name"
    fi
    echo "Reconfigured '$name' to port $newport"
}

# Clones an instance to a new name and port
cmd_clone() {
    local old="$1" new="$2" newport="$3"
    require_instance_exists "$old"
    cp -a "$(db_path "$old")" "$(db_path "$new")"
    echo "PORT=$newport" > "$(meta_file "$new")"
    if systemctl list-units --type=service | grep -q "$(service_name "$old")"; then
        cmd_setDefault "$new"
    fi
    echo "Cloned '$old' to '$new' (port $newport)"
}

# Creates a tar.gz backup of the instance
cmd_backup() {
    local name="$1"
    require_instance_exists "$name"
    local out="$BACKUP_DIR/${name}_$(date +%Y%m%d%H%M%S).tar.gz"
    tar -czf "$out" -C "$DB_ROOT" "$name" "$(basename "$(meta_file "$name")")"
    echo "Backup created: $out"
}

# Logs the activities of the instance
cmd_logs() {
    local name="$1"
    require_instance_exists "$name"
    local log="$(log_path "$name")"
    if [ -f "$log" ]; then
        echo "Tailing logs for '$name' (Ctrl+C to exit):"
        tail -f "$log"
    else
        echo "Log file not found for '$name'"
    fi
}

# Displays the info of the instance
cmd_info() {
    local name="$1"
    require_instance_exists "$name"
    local port="$(read_port "$name")"
    local dbp="$(db_path "$name")"
    local logp="$(log_path "$name")"
    local svc="$(service_name "$name")"
    local status="❌ Dead"
    if pgrep -f "mongod.*--port $port" > /dev/null; then
        status="✅ Running"
        elif systemctl is-enabled "$svc" &> /dev/null; then
        status="🔁 Enabled"
    fi
    
    echo "📋 Instance Info: $name"
    echo " - Port: $port"
    echo " - Data Path: $dbp"
    echo " - Log Path: $logp"
    echo " - Systemd Service: $svc"
    echo " - Status: $status"
}

# Displays the help and usage menu
cmd_help() {
    echo "Available options" 
echo      add name port -	Add new instance
echo      name -	Launch instance directly
echo      delete name -	Delete instance and service
echo      setDefault name -	Enable as systemd/Windows service 
echo      start/-stop/-restart name - 	Control systemd/Windows service
echo      kill name	Kill - direct process
echo      rename old new -	Rename instance and service
echo      reconfigure name port -	Change port and update service
echo      clone old new port -	Clone instance to new name/port
echo      backup name -	Create tar.gz or zip backup
echo      list -	List all instances
echo      status -	Show running/enabled/dead status
echo      logs n ame -	Tail log file live
echo      info name -	Show metadata and service status

}

# === DISPATCHER ===
case "${1:-}" in
    add)         cmd_add "$2" "$3" ;;
    delete)      cmd_delete "$2" ;;
    setDefault)  cmd_setDefault "$2" ;;
    list)        cmd_list ;;
    status)      cmd_status ;;
    rename)      cmd_rename "$2" "$3" ;;
    reconfigure) cmd_reconfigure "$2" "$3" ;;
    clone)       cmd_clone "$2" "$3" "$4" ;;
    backup)      cmd_backup "$2" ;;
    "")           if [ -n "${2:-}" ]; then cmd_launch "$2"; else echo "Usage: mongoInstance [options]"; fi ;;
    start)       cmd_start "$2" ;;
    stop)        cmd_stop "$2" ;;
    restart)     cmd_restart "$2" ;;
    kill)        cmd_kill "$2" ;;
    launch)      cmd_launch "$2" ;;
    logs)        cmd_logs "$2" ;;
    info)        cmd_info "$2" ;;
    *)           cmd_help ;;
esac
