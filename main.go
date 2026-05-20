// mongoman — Cross-platform MongoDB Instance Manager.
//
// A modular MongoDB instance manager for Linux, macOS, BSD, and Windows.
// Add, launch, clone, rename, reconfigure, backup, and control multiple
// MongoDB instances with ease—whether using direct process control or
// OS-native services (systemd, launchd, Windows Service).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/davisdeveloper/mongoman/internal/config"
	"github.com/davisdeveloper/mongoman/internal/instance"
	"github.com/davisdeveloper/mongoman/internal/proc"
	"github.com/davisdeveloper/mongoman/internal/service"
)

var (
	paths  *config.Paths
	svcMgr service.Manager
)

func main() {
	var err error
	paths, err = config.DefaultPaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing paths: %v\n", err)
		os.Exit(1)
	}

	svcMgr = service.NewManager()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	if err := dispatch(cmd, args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func dispatch(cmd string, args []string) error {
	switch cmd {
	case "add":
		return cmdAdd(args)
	case "launch":
		return cmdLaunch(args)
	case "kill":
		return cmdKill(args)
	case "delete":
		return cmdDelete(args)
	case "list":
		return cmdList()
	case "status":
		return cmdStatus()
	case "rename":
		return cmdRename(args)
	case "reconfigure":
		return cmdReconfigure(args)
	case "clone":
		return cmdClone(args)
	case "backup":
		return cmdBackup(args)
	case "logs":
		return cmdLogs(args)
	case "info":
		return cmdInfo(args)
	case "history":
		return cmdHistory(args)
	case "shell":
		return cmdShell(args)
	case "enable":
		return cmdEnable(args)
	case "disable":
		return cmdDisable(args)
	case "start":
		return cmdStart(args)
	case "stop":
		return cmdStop(args)
	case "restart":
		return cmdRestart(args)
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		// Backward-compatible: if the arg looks like a name (not a flag),
		// treat as "launch <name>".
		if !strings.HasPrefix(cmd, "-") {
			return cmdLaunch([]string{cmd})
		}
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

// ────────────────────────── COMMAND IMPLEMENTATIONS ──────────────────────────

// Usage: mongoman add <name> <port>
func cmdAdd(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: mongoman add <name> <port>")
	}
	name := args[0]
	port, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid port number: %s", args[1])
	}

	inst, err := instance.Create(paths, name, port)
	if err != nil {
		return err
	}
	fmt.Printf("✅ Added instance %q at %s (port %d)\n", name, inst.DBPath, port)
	return nil
}

// Usage: mongoman launch <name>
func cmdLaunch(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mongoman launch <name>")
	}
	name := args[0]
	inst := instance.RequireExists(paths, name)
	return proc.Launch(inst)
}

// Usage: mongoman kill <name>
func cmdKill(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mongoman kill <name>")
	}
	name := args[0]
	inst := instance.RequireExists(paths, name)
	return proc.Kill(inst)
}

// Usage: mongoman delete <name>
func cmdDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mongoman delete <name>")
	}
	name := args[0]
	return instance.Delete(paths, name)
}

// Usage: mongoman rename <old> <new>
func cmdRename(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: mongoman rename <old> <new>")
	}
	oldName, newName := args[0], args[1]

	if err := instance.Rename(paths, oldName, newName); err != nil {
		return err
	}

	// If service was enabled, recreate it with new name.
	if svcMgr.IsEnabled(oldName) {
		oldInst := instance.RequireExists(paths, newName)
		svcMgr.Disable(oldName)
		svcMgr.Enable(newName, oldInst.Meta.Port, oldInst.DBPath, oldInst.LogPath)
	}

	fmt.Printf("Renamed %q to %q\n", oldName, newName)
	return nil
}

// Usage: mongoman reconfigure <name> <new-port>
func cmdReconfigure(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: mongoman reconfigure <name> <new-port>")
	}
	name := args[0]
	newPort, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid port number: %s", args[1])
	}

	if err := instance.Reconfigure(paths, name, newPort); err != nil {
		return err
	}

	// Update service if enabled.
	if svcMgr.IsEnabled(name) {
		inst := instance.RequireExists(paths, name)
		svcMgr.Disable(name)
		svcMgr.Enable(name, inst.Meta.Port, inst.DBPath, inst.LogPath)
	}

	fmt.Printf("Reconfigured %q to port %d\n", name, newPort)
	return nil
}

// Usage: mongoman clone <src> <dst> <port>
func cmdClone(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: mongoman clone <src> <dst> <port>")
	}
	srcName, dstName := args[0], args[1]
	port, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid port number: %s", args[2])
	}

	srcInst := instance.RequireExists(paths, srcName)
	srcDB := paths.DBPath(srcName)
	dstDB := paths.DBPath(dstName)

	// Copy data directory.
	if err := copyDir(srcDB, dstDB); err != nil {
		return fmt.Errorf("cannot clone data directory: %w", err)
	}

	// Create new metadata.
	_, err = instance.Create(paths, dstName, port)
	if err != nil {
		return err
	}

	// Clone service if original has one.
	if svcMgr.IsEnabled(srcName) {
		svcMgr.Enable(dstName, port, dstDB, paths.LogPath(dstName))
	}

	fmt.Printf("Cloned %q to %q (port %d)\n", srcName, dstName, port)
	_ = srcInst // used for validation above
	return nil
}

// Usage: mongoman backup <name>
func cmdBackup(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mongoman backup <name>")
	}
	name := args[0]
	inst := instance.RequireExists(paths, name)

	timestamp := time.Now().Format("20060102150405")
	backupPath := paths.BackupPath(name, timestamp)

	// Create a tar.gz of the data directory.
	cmd := exec.Command("tar", "-czf", backupPath, "-C", paths.DataDir, name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	fmt.Printf("✅ Backup created: %s\n", backupPath)
	_ = inst
	return nil
}

// Usage: mongoman list
func cmdList() error {
	names, err := instance.ListAll(paths)
	if err != nil {
		return err
	}

	if len(names) == 0 {
		fmt.Println("📦 No MongoDB instances found.")
		return nil
	}

	sort.Strings(names)
	fmt.Println("📦 MongoDB Instances:")
	for _, name := range names {
		inst, err := instance.Load(paths, name)
		if err != nil {
			fmt.Printf(" - %s (error loading: %v)\n", name, err)
			continue
		}
		fmt.Printf(" - %s (port: %d)\n", name, inst.Meta.Port)
	}
	return nil
}

// Usage: mongoman status
func cmdStatus() error {
	names, err := instance.ListAll(paths)
	if err != nil {
		return err
	}

	if len(names) == 0 {
		fmt.Println("📊 No MongoDB instances found.")
		return nil
	}

	sort.Strings(names)
	fmt.Println("📊 MongoDB Instance Status:")
	for _, name := range names {
		inst, err := instance.Load(paths, name)
		if err != nil {
			fmt.Printf(" - %s (error: %v)\n", name, err)
			continue
		}

		status := svcMgr.Status(name)
		if status == service.StatusStopped {
			if proc.IsRunning(inst) {
				status = service.StatusRunning
			}
		}

		fmt.Printf(" - %s (port: %d): %s\n", name, inst.Meta.Port, status)
	}
	return nil
}

// Usage: mongoman logs <name>
func cmdLogs(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mongoman logs <name>")
	}
	name := args[0]
	_ = instance.RequireExists(paths, name)

	logPath := paths.LogPath(name)
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return fmt.Errorf("log file not found for %q", name)
	}

	fmt.Printf("📝 Tailing logs for %q (Ctrl+C to exit):\n", name)
	cmd := exec.Command("tail", "-f", logPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Usage: mongoman info <name>
func cmdInfo(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mongoman info <name>")
	}
	name := args[0]
	inst := instance.RequireExists(paths, name)

	status := svcMgr.Status(name)
	if status == service.StatusStopped {
		if proc.IsRunning(inst) {
			status = service.StatusRunning
		}
	}

	fmt.Printf("📋 Instance Info: %s\n", name)
	fmt.Printf("   Port:          %d\n", inst.Meta.Port)
	fmt.Printf("   Data Path:     %s\n", inst.DBPath)
	fmt.Printf("   Log Path:      %s\n", inst.LogPath)
	fmt.Printf("   Created:       %s\n", inst.Meta.CreatedAt.Format(time.RFC3339))
	fmt.Printf("   Launch Count:  %d\n", inst.Meta.LaunchCount)
	fmt.Printf("   Status:        %s\n", status)
	return nil
}

// Usage: mongoman history <name>
func cmdHistory(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mongoman history <name>")
	}
	name := args[0]
	inst := instance.RequireExists(paths, name)

	data, err := json.MarshalIndent(inst.Meta, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot format history: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// Usage: mongoman shell <name>
func cmdShell(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mongoman shell <name>")
	}
	name := args[0]
	inst := instance.RequireExists(paths, name)

	if !proc.IsRunning(inst) {
		return fmt.Errorf("instance %q is not running. Launch it first", name)
	}

	port := strconv.Itoa(inst.Meta.Port)
	fmt.Printf("🚀 Launching mongosh for %q (port %s)\n", name, port)
	cmd := exec.Command("mongosh", "--port", port)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ────────────────────────── SERVICE COMMANDS ─────────────────────────────────

// Usage: mongoman enable <name>
func cmdEnable(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mongoman enable <name>")
	}
	name := args[0]
	inst := instance.RequireExists(paths, name)
	return svcMgr.Enable(name, inst.Meta.Port, inst.DBPath, inst.LogPath)
}

// Usage: mongoman disable <name>
func cmdDisable(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mongoman disable <name>")
	}
	name := args[0]
	return svcMgr.Disable(name)
}

// Usage: mongoman start <name>
func cmdStart(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mongoman start <name>")
	}
	name := args[0]
	return svcMgr.Start(name)
}

// Usage: mongoman stop <name>
func cmdStop(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mongoman stop <name>")
	}
	name := args[0]
	return svcMgr.Stop(name)
}

// Usage: mongoman restart <name>
func cmdRestart(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mongoman restart <name>")
	}
	name := args[0]
	return svcMgr.Restart(name)
}

// ────────────────────────── HELPERS ──────────────────────────────────────────

func printUsage() {
	fmt.Println(`mongoman — Cross-platform MongoDB Instance Manager

Usage:
  mongoman add <name> <port>          Add a new MongoDB instance
  mongoman launch <name>              Launch instance (forked process)
  mongoman kill <name>                Kill direct process
  mongoman delete <name>              Delete instance
  mongoman list                       List all instances
  mongoman status                     Show running/enabled/dead status
  mongoman rename <old> <new>         Rename instance
  mongoman reconfigure <name> <port>  Change port
  mongoman clone <src> <dst> <port>   Clone instance
  mongoman backup <name>              Create backup archive
  mongoman logs <name>                Tail log file
  mongoman info <name>                Show metadata and status
  mongoman history <name>             Show launch history
  mongoman shell <name>               Launch mongosh for instance
  mongoman enable <name>              Enable as OS service
  mongoman disable <name>             Disable OS service
  mongoman start <name>               Start OS service
  mongoman stop <name>                Stop OS service
  mongoman restart <name>             Restart OS service
  mongoman help                       Show this help message

Directory Layout:
  Data:     ~/mongoman/data/<name>
  Logs:     ~/mongoman/logs/<name>.log
  Backups:  ~/mongoman/backups/<name>_<timestamp>.*
  Config:   ~/.config/mongoman/<name>.json  (or %%APPDATA%% on Windows)`)
}

// copyDir copies a directory recursively using platform-native commands.
func copyDir(src, dst string) error {
	var cmd *exec.Cmd
	// Use cp on Unix, robocopy on Windows.
	if isWindows() {
		cmd = exec.Command("robocopy", src, dst, "/E", "/NFL", "/NDL", "/NJH", "/NJS")
	} else {
		cmd = exec.Command("cp", "-a", src, dst)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func isWindows() bool {
	return len(os.Getenv("SYSTEMROOT")) > 0 && strings.HasPrefix(os.Getenv("SYSTEMROOT"), "C:\\")
}
