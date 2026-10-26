// Package backup creates compressed archives of MongoDB instance data.
//
// Pure-Go implementation — no external tar/zip CLI required, so backups work
// identically on Linux, macOS, BSD, and Windows.
package backup

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Format selects the archive backend.
type Format int

const (
	FormatTarGz Format = iota // tar.gz (Unix)
	FormatZip                 // zip (Windows)
)

// Create archives the instance data directory (dbPath) into dst using the
// given format. Archive entries are rooted at name (e.g. "myinst/...").
func Create(dst, dbPath, name string, format Format) error {
	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("cannot create archive %s: %w", dst, err)
	}
	defer f.Close()

	if _, err := os.Stat(dbPath); err != nil {
		return fmt.Errorf("instance data dir %q not found: %w", dbPath, err)
	}

	switch format {
	case FormatZip:
		return writeZip(f, dbPath, name)
	default:
		return writeTarGz(f, dbPath, name)
	}
}

// writeTarGz writes a gzipped tar archive with entries rooted at name.
func writeTarGz(w io.Writer, src, name string) error {
	gw := gzip.NewWriter(w)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return addTarEntry(tw, path, src, name, info)
	})
}

// writeZip writes a zip archive with entries rooted at name.
func writeZip(w io.Writer, src, name string) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return addZipEntry(zw, path, src, name, info)
	})
}

// archivePath computes the path stored inside the archive, rooted at name.
func archivePath(src, name, path string) (string, error) {
	rel, err := filepath.Rel(src, path)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Join(name, rel)), nil
}

// addTarEntry writes a single file/directory into the tar writer.
func addTarEntry(tw *tar.Writer, path, src, name string, info os.FileInfo) error {
	arc, err := archivePath(src, name, path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return tw.WriteHeader(&tar.Header{
			Name:     arc + "/",
			Mode:     int64(info.Mode()),
			Typeflag: tar.TypeDir,
			ModTime:  info.ModTime(),
		})
	}

	hdr := &tar.Header{
		Name:     arc,
		Mode:     int64(info.Mode()),
		Size:     info.Size(),
		Typeflag: tar.TypeReg,
		ModTime:  info.ModTime(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	_, err = io.Copy(tw, in)
	return err
}

// addZipEntry writes a single file/directory into the zip writer.
func addZipEntry(zw *zip.Writer, path, src, name string, info os.FileInfo) error {
	arc, err := archivePath(src, name, path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		arc += "/"
		hdr := &zip.FileHeader{
			Name:     arc,
			Modified: info.ModTime(),
		}
		hdr.SetMode(os.ModeDir | 0o755)
		_, err := zw.CreateHeader(hdr)
		return err
	}

	hdr := &zip.FileHeader{
		Name:     arc,
		Modified: info.ModTime(),
	}
	hdr.SetMode(info.Mode())
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	_, err = io.Copy(w, in)
	return err
}
