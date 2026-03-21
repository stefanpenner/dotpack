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

	ext := "tar.gz"
	if osName == "windows" {
		ext = "zip"
	}

	bundle := filepath.Join(scriptDir, fmt.Sprintf("dotpack-%s-%s.%s", osName, arch, ext))
	if _, err := os.Stat(bundle); os.IsNotExist(err) {
		return fmt.Errorf("no bundle found at %s\nRun: dotpack build --os %s --arch %s", bundle, osName, arch)
	}

	prefix := defaultPrefix()

	fmt.Printf("==> Installing to %s...\n", prefix)
	if err := os.MkdirAll(prefix, 0755); err != nil {
		return err
	}

	var err error
	if ext == "zip" {
		err = archive.ExtractZip(bundle, prefix)
	} else {
		err = archive.ExtractTarGz(bundle, prefix)
	}
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}
	fmt.Printf("==> Done. Ensure PATH includes %s%cbin\n", prefix, filepath.Separator)
	return nil
}

// defaultPrefix returns the install location from DOTPACK_PREFIX or the platform default.
func defaultPrefix() string {
	if p := os.Getenv("DOTPACK_PREFIX"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "dotpack")
		}
		return filepath.Join(home, "AppData", "Local", "dotpack")
	}
	return filepath.Join(home, ".local")
}
