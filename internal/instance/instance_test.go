package instance

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davisdeveloper/mongoman/internal/config"
)

func newTestPaths(t *testing.T) *config.Paths {
	t.Helper()
	dir := t.TempDir()
	// Override the default paths to use our temp directory
	p := &config.Paths{
		DataDir:    filepath.Join(dir, "data"),
		LogsDir:    filepath.Join(dir, "logs"),
		BackupsDir: filepath.Join(dir, "backups"),
		ConfigDir:  filepath.Join(dir, "config"),
	}
	// Create required directories (Create() expects them to exist)
	for _, d := range []string{p.DataDir, p.LogsDir, p.ConfigDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	return p
}

func TestCreateLoad(t *testing.T) {
	paths := newTestPaths(t)

	inst, err := Create(paths, "testinst", 27017)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if inst.Meta.Name != "testinst" {
		t.Errorf("name = %q, want %q", inst.Meta.Name, "testinst")
	}
	if inst.Meta.Port != 27017 {
		t.Errorf("port = %d, want 27017", inst.Meta.Port)
	}
	if inst.Meta.LaunchCount != 0 {
		t.Errorf("launch_count = %d, want 0", inst.Meta.LaunchCount)
	}
	if inst.Meta.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if inst.Meta.DeletedAt != nil {
		t.Error("DeletedAt should be nil")
	}

	// Verify files exist on disk
	if _, err := os.Stat(inst.DBPath); err != nil {
		t.Errorf("data dir not created: %v", err)
	}
	if _, err := os.Stat(inst.LogPath); err != nil {
		t.Errorf("log file not created: %v", err)
	}
	if _, err := os.Stat(inst.MetaPath); err != nil {
		t.Errorf("metadata file not created: %v", err)
	}

	// Load it back
	loaded, err := Load(paths, "testinst")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Meta.Name != inst.Meta.Name {
		t.Errorf("loaded name mismatch")
	}
	if loaded.Meta.Port != inst.Meta.Port {
		t.Errorf("loaded port mismatch")
	}
	// JSON marshaling loses sub-second precision; compare within 1 second
	diff := loaded.Meta.CreatedAt.Sub(inst.Meta.CreatedAt)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("loaded CreatedAt mismatch: %v vs %v", loaded.Meta.CreatedAt, inst.Meta.CreatedAt)
	}
}

func TestCreateDuplicate(t *testing.T) {
	paths := newTestPaths(t)
	if _, err := Create(paths, "dup", 27017); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if _, err := Create(paths, "dup", 27018); err == nil {
		t.Fatal("second Create should fail with duplicate error")
	}
}

func TestLoadMissing(t *testing.T) {
	paths := newTestPaths(t)
	if _, err := Load(paths, "nonexistent"); err == nil {
		t.Fatal("Load should error on missing instance")
	}
}

func TestRequireExists(t *testing.T) {
	paths := newTestPaths(t)
	if _, err := RequireExists(paths, "missing"); err == nil {
		t.Fatal("RequireExists should error on missing instance")
	}
	if _, err := Create(paths, "exists", 27017); err != nil {
		t.Fatalf("Create: %v", err)
	}
	inst, err := RequireExists(paths, "exists")
	if err != nil {
		t.Fatalf("RequireExists on existing: %v", err)
	}
	if inst.Meta.Name != "exists" {
		t.Errorf("wrong instance returned")
	}
}

func TestReconfigure(t *testing.T) {
	paths := newTestPaths(t)
	if _, err := Create(paths, "reconf", 27017); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := Reconfigure(paths, "reconf", 27018); err != nil {
		t.Fatalf("Reconfigure: %v", err)
	}
	inst, err := Load(paths, "reconf")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if inst.Meta.Port != 27018 {
		t.Errorf("port not updated: %d", inst.Meta.Port)
	}
}

func TestRename(t *testing.T) {
	paths := newTestPaths(t)
	if _, err := Create(paths, "oldname", 27017); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := Rename(paths, "oldname", "newname"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	// Old should be gone
	if Exists(paths, "oldname") {
		t.Error("old name still exists")
	}
	// New should exist
	inst, err := Load(paths, "newname")
	if err != nil {
		t.Fatalf("Load new: %v", err)
	}
	if inst.Meta.Name != "newname" {
		t.Errorf("meta name not updated: %s", inst.Meta.Name)
	}
	// Data dir moved
	if _, err := os.Stat(paths.DBPath("oldname")); !os.IsNotExist(err) {
		t.Error("old data dir still exists")
	}
	if _, err := os.Stat(paths.DBPath("newname")); err != nil {
		t.Errorf("new data dir not found: %v", err)
	}
	// Metadata file moved
	if _, err := os.Stat(paths.MetaFile("oldname")); !os.IsNotExist(err) {
		t.Error("old meta file still exists")
	}
	if _, err := os.Stat(paths.MetaFile("newname")); err != nil {
		t.Errorf("new meta file not found: %v", err)
	}
}

func TestDelete(t *testing.T) {
	paths := newTestPaths(t)
	if _, err := Create(paths, "todelete", 27017); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := Delete(paths, "todelete"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if Exists(paths, "todelete") {
		t.Error("instance still exists after delete")
	}
	if _, err := os.Stat(paths.DBPath("todelete")); !os.IsNotExist(err) {
		t.Error("data dir still exists after delete")
	}
	if _, err := os.Stat(paths.MetaFile("todelete")); !os.IsNotExist(err) {
		t.Error("meta file still exists after delete")
	}
}

func TestRecordLaunchAndKill(t *testing.T) {
	paths := newTestPaths(t)
	inst, err := Create(paths, "launchkill", 27017)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	before := time.Now()
	if err := inst.RecordLaunch(); err != nil {
		t.Fatalf("RecordLaunch: %v", err)
	}
	if inst.Meta.LaunchCount != 1 {
		t.Errorf("LaunchCount = %d, want 1", inst.Meta.LaunchCount)
	}
	if len(inst.Meta.LaunchHistory) != 1 {
		t.Fatalf("LaunchHistory len = %d, want 1", len(inst.Meta.LaunchHistory))
	}
	if inst.Meta.LaunchHistory[0].Start.Before(before) {
		t.Error("Launch start time before call")
	}
	if inst.Meta.LaunchHistory[0].End != nil {
		t.Error("Launch end should be nil")
	}

	if err := inst.RecordKill(); err != nil {
		t.Fatalf("RecordKill: %v", err)
	}
	if inst.Meta.LaunchHistory[0].End == nil {
		t.Error("Launch end should be set after RecordKill")
	}
	if inst.Meta.LaunchHistory[0].End.Before(before) {
		t.Error("Launch end time before call")
	}
}

func TestListAll(t *testing.T) {
	paths := newTestPaths(t)
	if _, err := Create(paths, "a", 27017); err != nil {
		t.Fatalf("Create a: %v", err)
	}
	if _, err := Create(paths, "b", 27018); err != nil {
		t.Fatalf("Create b: %v", err)
	}
	names, err := ListAll(paths)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("ListAll returned %d names, want 2", len(names))
	}
	// Should be sorted by ListAll
	if names[0] != "a" || names[1] != "b" {
		t.Errorf("names = %v, want [a b]", names)
	}
}
