package download

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// File downloads a URL to a local file path.
func File(url, dest string) error {
	fmt.Printf("  %s\n", filepath.Base(dest))
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// TarGzBinary downloads a .tar.gz and extracts a single named binary.
func TarGzBinary(url, outputDir, binaryName string) error {
	fmt.Printf("  %s\n", binaryName)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("gzip %s: %w", url, err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("%s not found in %s", binaryName, url)
		}
		if err != nil {
			return fmt.Errorf("tar %s: %w", url, err)
		}
		if filepath.Base(hdr.Name) == binaryName && hdr.Typeflag == tar.TypeReg {
			dest := filepath.Join(outputDir, binaryName)
			f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			_, err = io.Copy(f, tr)
			f.Close()
			return err
		}
	}
}

// TarGzFull downloads a .tar.gz and extracts everything to outputDir.
// If stripComponents > 0, it strips that many leading path components.
func TarGzFull(url, outputDir string, stripComponents int) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar %s: %w", url, err)
		}

		name := hdr.Name
		if stripComponents > 0 {
			parts := strings.SplitN(name, "/", stripComponents+1)
			if len(parts) <= stripComponents {
				continue
			}
			name = parts[stripComponents]
			if name == "" {
				continue
			}
		}

		target := filepath.Join(outputDir, name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)|0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			_, err = io.Copy(f, tr)
			f.Close()
			if err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			os.Remove(target) // remove existing if any
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return err
			}
		}
	}
	return nil
}

// ZipFiles downloads a .zip and extracts specified files.
// fileMap maps archive paths (relative) to destination paths (absolute).
func ZipFiles(url string, fileMap map[string]string) error {
	// Zip requires seeking, so download to temp file first
	tmp, err := os.CreateTemp("", "dotpack-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	size, err := io.Copy(tmp, resp.Body)
	if err != nil {
		return err
	}

	r, err := zip.NewReader(tmp, size)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		dest, ok := fileMap[f.Name]
		if !ok {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			rc.Close()
			return err
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode()|0755)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// TarGzToDir downloads a .tar.gz, extracts it, and renames the top-level
// directory to destDir. Used for plugins where the archive has a single
// top-level directory like "repo-name-version/".
func TarGzToDir(url, destDir string) error {
	tmp, err := os.MkdirTemp("", "dotpack-plugin-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	if err := TarGzFull(url, tmp, 0); err != nil {
		return err
	}

	// Find the single top-level directory
	entries, err := os.ReadDir(tmp)
	if err != nil {
		return err
	}
	if len(entries) != 1 || !entries[0].IsDir() {
		// Fallback: just move everything
		return os.Rename(tmp, destDir)
	}
	return os.Rename(filepath.Join(tmp, entries[0].Name()), destDir)
}
