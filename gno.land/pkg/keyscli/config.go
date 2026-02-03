package keyscli

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

const configFile = "config.toml"

// Config represents the gnokey configuration file.
// Supports global settings and per-zone overrides.
type Config struct {
	Global GlobalConfig          `toml:"global,omitempty"`
	Zones  map[string]ZoneConfig `toml:"zones,omitempty"`
}

// GlobalConfig holds settings that apply to all zones unless overridden.
type GlobalConfig struct {
	// Future global settings can be added here
}

// ZoneConfig holds settings specific to a remote/zone.
type ZoneConfig struct {
	CLAHash string `toml:"cla_hash,omitempty"`
	// Future per-zone settings can be added here
}

// LoadConfig reads the config file from the gnokey home directory.
// Returns empty config if file doesn't exist.
func LoadConfig(home string) (*Config, error) {
	path := filepath.Join(home, configFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Zones: make(map[string]ZoneConfig)}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Zones == nil {
		cfg.Zones = make(map[string]ZoneConfig)
	}
	return &cfg, nil
}

// SaveConfig writes the config file to the gnokey home directory.
func SaveConfig(home string, cfg *Config) error {
	path := filepath.Join(home, configFile)
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// GetCLAHash returns the CLA hash for a specific remote.
// Returns empty string if not found.
func (c *Config) GetCLAHash(remote string) string {
	if zone, ok := c.Zones[remote]; ok {
		return zone.CLAHash
	}
	return ""
}

// SetCLAHash sets the CLA hash for a specific remote.
func (c *Config) SetCLAHash(remote, hash string) {
	if c.Zones == nil {
		c.Zones = make(map[string]ZoneConfig)
	}
	zone := c.Zones[remote]
	zone.CLAHash = hash
	c.Zones[remote] = zone
}
