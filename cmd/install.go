package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/stefanpenner/dotpack/internal/archive"
)

// Install extracts a dotpack bundle to the local DOTPACK_PREFIX.
func Install(scriptDir string) error {
	osName := strings.ToLower(runtime.GOOS)
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		if osName == "linux" {
			arch = "aarch64"
		}
	}

	tarball := filepath.Join(scriptDir, fmt.Sprintf("dotpack-%s-%s.tar.gz", osName, arch))
	if _, err := os.Stat(tarball); os.IsNotExist(err) {
		return fmt.Errorf("no bundle found at %s\nRun: dotpack build --os %s --arch %s", tarball, osName, arch)
	}

	prefix := os.Getenv("DOTPACK_PREFIX")
	if prefix == "" {
		home, _ := os.UserHomeDir()
		prefix = filepath.Join(home, ".local")
	}

	fmt.Printf("==> Installing to %s...\n", prefix)
	if err := os.MkdirAll(prefix, 0755); err != nil {
		return err
	}
	if err := archive.ExtractTarGz(tarball, prefix); err != nil {
		return fmt.Errorf("extract: %w", err)
	}
	fmt.Printf("==> Done. Ensure PATH includes %s/bin\n", prefix)
	return nil
}
