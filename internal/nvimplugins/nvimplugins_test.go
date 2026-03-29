package nvimplugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLockfile(t *testing.T) {
	dir := t.TempDir()
	lockfile := filepath.Join(dir, "nvim-pack-lock.json")
	os.WriteFile(lockfile, []byte(`{
  "telescope.nvim": { "branch": "master", "commit": "abc123" },
  "nvim-treesitter": { "branch": "master", "commit": "def456" }
}`), 0644)

	plugins, err := ParseLockfile(lockfile)
	if err != nil {
		t.Fatalf("ParseLockfile: %v", err)
	}
	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}

	byName := map[string]Plugin{}
	for _, p := range plugins {
		byName[p.Name] = p
	}

	if p, ok := byName["telescope.nvim"]; !ok {
		t.Error("missing telescope.nvim")
	} else if p.Commit != "abc123" {
		t.Errorf("telescope commit = %s, want abc123", p.Commit)
	}
}

func TestSyncPlugins(t *testing.T) {
	dir := t.TempDir()

	// Create lockfile
	lockfile := filepath.Join(dir, "nvim-pack-lock.json")
	os.WriteFile(lockfile, []byte(`{
  "myplugin": { "branch": "main", "commit": "aaa" },
  "missing": { "branch": "main", "commit": "bbb" }
}`), 0644)

	// Create local plugin directory (only myplugin exists)
	localPlugins := filepath.Join(dir, "local-plugins")
	pluginDir := filepath.Join(localPlugins, "myplugin")
	os.MkdirAll(filepath.Join(pluginDir, ".git"), 0755)
	os.WriteFile(filepath.Join(pluginDir, "init.lua"), []byte("-- plugin"), 0644)
	os.WriteFile(filepath.Join(pluginDir, ".git", "config"), []byte("gitconfig"), 0644)

	// Sync
	destDir := filepath.Join(dir, "dest")
	err := SyncPlugins(lockfile, localPlugins, destDir)
	if err != nil {
		t.Fatalf("SyncPlugins: %v", err)
	}

	// Verify plugin was copied
	if _, err := os.Stat(filepath.Join(destDir, "myplugin", "init.lua")); err != nil {
		t.Error("init.lua not copied")
	}

	// Verify .git was excluded
	if _, err := os.Stat(filepath.Join(destDir, "myplugin", ".git")); !os.IsNotExist(err) {
		t.Error(".git directory should be excluded")
	}

	// missing plugin should be skipped (no error)
	if _, err := os.Stat(filepath.Join(destDir, "missing")); !os.IsNotExist(err) {
		t.Error("missing plugin should not exist in dest")
	}
}
