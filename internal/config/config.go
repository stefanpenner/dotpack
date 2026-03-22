package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the devlayer configuration.
type Config struct {
	Dotfiles DotfilesConfig `toml:"dotfiles"`
}

// DotfilesConfig lists paths (relative to $HOME) to sync.
type DotfilesConfig struct {
	Sync []string `toml:"sync"`
}

// Load reads the config file, returning empty defaults if it doesn't exist.
func Load() (*Config, error) {
	cfg := &Config{}
	path := Path()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Path returns the config file location.
func Path() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "devlayer", "config.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "devlayer", "config.toml")
}
