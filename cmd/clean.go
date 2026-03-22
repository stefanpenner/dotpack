package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/stefanpenner/devlayer/internal/docker"
)

// Clean removes build artifacts and Docker images.
func Clean(scriptDir string) error {
	fmt.Println("==> Cleaning up...")

	matches, _ := filepath.Glob(filepath.Join(scriptDir, "devlayer-*.tar.gz"))
	zipMatches, _ := filepath.Glob(filepath.Join(scriptDir, "devlayer-*.zip"))
	matches = append(matches, zipMatches...)
	for _, m := range matches {
		os.Remove(m)
	}
	if len(matches) > 0 {
		fmt.Printf("  Removed %d tarball(s)\n", len(matches))
	}

	if docker.ImageExists("devlayer") {
		docker.RemoveImage("devlayer")
		fmt.Println("  Removed Docker image")
	}

	fmt.Println("==> Done.")
	return nil
}
