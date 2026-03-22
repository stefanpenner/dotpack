package download

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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

		// Skip macOS AppleDouble resource fork files
		if strings.HasPrefix(filepath.Base(name), "._") {
			continue
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

// ZipBinary downloads a .zip and extracts a single named binary.
func ZipBinary(url, outputDir, binaryName string) error {
	fmt.Printf("  %s\n", binaryName)

	tmp, err := downloadToTemp(url, "devlayer-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	info, err := os.Stat(tmp)
	if err != nil {
		return err
	}

	f, err := os.Open(tmp)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := zip.NewReader(f, info.Size())
	if err != nil {
		return err
	}

	for _, zf := range r.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(zf.Name) == binaryName {
			rc, err := zf.Open()
			if err != nil {
				return err
			}
			dest := filepath.Join(outputDir, binaryName)
			out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				rc.Close()
				return err
			}
			_, err = io.Copy(out, rc)
			rc.Close()
			out.Close()
			return err
		}
	}
	return fmt.Errorf("%s not found in %s", binaryName, url)
}

// ZipFull downloads a .zip and extracts everything to outputDir.
// If stripComponents > 0, it strips that many leading path components.
func ZipFull(url, outputDir string, stripComponents int) error {
	tmp, err := downloadToTemp(url, "devlayer-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	info, err := os.Stat(tmp)
	if err != nil {
		return err
	}

	f, err := os.Open(tmp)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := zip.NewReader(f, info.Size())
	if err != nil {
		return err
	}

	for _, zf := range r.File {
		name := zf.Name
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

		if zf.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		rc, err := zf.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, zf.Mode()|0755)
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

// ZipFiles downloads a .zip and extracts specified files.
// fileMap maps archive paths (relative) to destination paths (absolute).
func ZipFiles(url string, fileMap map[string]string) error {
	tmp, err := downloadToTemp(url, "devlayer-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	info, err := os.Stat(tmp)
	if err != nil {
		return err
	}

	f, err := os.Open(tmp)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := zip.NewReader(f, info.Size())
	if err != nil {
		return err
	}

	for _, zf := range r.File {
		dest, ok := fileMap[zf.Name]
		if !ok {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			rc.Close()
			return err
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, zf.Mode()|0755)
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
	tmp, err := os.MkdirTemp("", "devlayer-plugin-*")
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

// TarXzFull downloads a .tar.xz and extracts everything to outputDir.
// If stripComponents > 0, it strips that many leading path components.
// Uses the system tar command since Go stdlib doesn't support xz.
func TarXzFull(url, outputDir string, stripComponents int) error {
	tmp, err := downloadToTemp(url, "devlayer-*.tar.xz")
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	args := []string{"xJf", tmp, "-C", outputDir}
	if stripComponents > 0 {
		args = append(args, fmt.Sprintf("--strip-components=%d", stripComponents))
	}

	cmd := exec.Command("tar", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// downloadToTemp downloads a URL to a temporary file and returns the path.
func downloadToTemp(url, pattern string) (string, error) {
	tmp, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}

	resp, err := http.Get(url)
	if err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	_, err = io.Copy(tmp, resp.Body)
	tmp.Close()
	if err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return tmp.Name(), nil
}
