package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Dotfiles.Sync) != 0 {
		t.Errorf("expected empty sync list, got %v", cfg.Dotfiles.Sync)
	}
}

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "devlayer")
	os.MkdirAll(cfgDir, 0755)
	os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(`
[dotfiles]
sync = [
  ".config/nvim",
  ".zshrc",
]
`), 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Dotfiles.Sync) != 2 {
		t.Fatalf("expected 2 sync entries, got %d", len(cfg.Dotfiles.Sync))
	}
	if cfg.Dotfiles.Sync[0] != ".config/nvim" {
		t.Errorf("expected .config/nvim, got %s", cfg.Dotfiles.Sync[0])
	}
}
