package nvimplugins

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Plugin represents one entry from lazy-lock.json.
type Plugin struct {
	Name   string
	Branch string `json:"branch"`
	Commit string `json:"commit"`
}

// ParseLazyLock reads lazy-lock.json and returns plugin entries.
func ParseLazyLock(path string) ([]Plugin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw map[string]Plugin
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse lazy-lock.json: %w", err)
	}

	plugins := make([]Plugin, 0, len(raw))
	for name, p := range raw {
		p.Name = name
		plugins = append(plugins, p)
	}
	return plugins, nil
}

// LazyLockPath returns the default lazy-lock.json location.
func LazyLockPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "nvim", "lazy-lock.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "nvim", "lazy-lock.json")
}

// LocalLazyDir returns the default lazy plugin install directory.
func LocalLazyDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "nvim", "lazy")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "nvim", "lazy")
}

// SyncPlugins copies locally-installed plugins (at their lazy-lock.json pinned
// commits) into destDir, stripping .git directories to save space.
func SyncPlugins(lockfilePath, localLazyDir, destDir string) error {
	plugins, err := ParseLazyLock(lockfilePath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	var skipped []string
	for _, p := range plugins {
		srcDir := filepath.Join(localLazyDir, p.Name)
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			skipped = append(skipped, p.Name)
			continue
		}

		dst := filepath.Join(destDir, p.Name)
		if err := copyDirNoGit(srcDir, dst); err != nil {
			return fmt.Errorf("copy plugin %s: %w", p.Name, err)
		}
	}

	if len(skipped) > 0 {
		fmt.Printf("  skipped %d plugins not found locally: %s\n", len(skipped), strings.Join(skipped, ", "))
	}

	return nil
}

// copyDirNoGit recursively copies src to dst, skipping .git directories.
func copyDirNoGit(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip .git directories
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		// Skip symlinks to avoid issues
		if d.Type()&fs.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return nil // skip broken symlinks
			}
			return os.Symlink(linkTarget, target)
		}

		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
