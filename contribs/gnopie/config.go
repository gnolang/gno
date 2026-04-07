package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"
)

// Config holds persisted gnopie settings.
type Config struct {
	Key string `toml:"key,omitempty"` // default key name
	// Future: other defaults like gas-fee, quiet, etc.
}

func configPath(home string) string {
	return filepath.Join(home, "gnopie", "config.toml")
}

func LoadConfig(home string) (*Config, error) {
	path := configPath(home)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

func SaveConfig(home string, cfg *Config) error {
	path := configPath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ConfigGet returns the value for a known config key.
func ConfigGet(cfg *Config, key string) (string, error) {
	switch key {
	case "key":
		return cfg.Key, nil
	default:
		return "", fmt.Errorf("unknown config key %q (available: key)", key)
	}
}

// ConfigSet sets a config key to a value.
func ConfigSet(cfg *Config, key, value string) error {
	switch key {
	case "key":
		cfg.Key = value
		return nil
	default:
		return fmt.Errorf("unknown config key %q (available: key)", key)
	}
}

// ConfigList returns all config key=value pairs.
func ConfigList(cfg *Config) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "key=%s\n", cfg.Key)
	return sb.String()
}
