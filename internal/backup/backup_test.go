package backup

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateTarGz(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "data", "myinst")
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "file1.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "file2.txt"), []byte("world"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(dir, "out.tar.gz")
	if err := Create(out, filepath.Join(dir, "data"), "myinst", FormatTarGz); err != nil {
		t.Fatalf("Create: %v", err)
	}

	entries := readTarGz(t, out)
	want := map[string]bool{"myinst/file1.txt": false, "myinst/sub/file2.txt": false, "myinst/sub/": true, "myinst/": true}
	for _, e := range entries {
		if _, ok := want[e.name]; ok {
			want[e.name] = true
		} else {
			t.Errorf("unexpected entry %q", e.name)
		}
	}
	for k, v := range want {
		if !v {
			t.Errorf("missing entry %q", k)
		}
	}
	if len(entries) != len(want) {
		t.Errorf("expected %d entries, got %d", len(want), len(entries))
	}
}

func TestCreateZip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "data", "myinst")
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "file1.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(dir, "out.zip")
	if err := Create(out, filepath.Join(dir, "data"), "myinst", FormatZip); err != nil {
		t.Fatalf("Create: %v", err)
	}

	r, err := zip.OpenReader(out)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if len(r.File) < 2 {
		t.Fatalf("expected at least 2 entries, got %d", len(r.File))
	}
}

type tarEntry struct {
	name string
}

func readTarGz(t *testing.T, path string) []tarEntry {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	g, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	tr := tar.NewReader(g)
	var entries []tarEntry
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, tarEntry{name: h.Name})
	}
	return entries
}
