package gnomod

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

// TomlFile represents the structure of gnomod.toml
type TomlFile struct {
	Module struct {
		Path    string `toml:"path"`
		Version string `toml:"version"`
	} `toml:"module"`
	Draft bool `toml:"draft,omitempty"`
}

// ParseTomlFile parses a gnomod.toml file
func ParseTomlFile(fpath string) (*File, error) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	var f File
	if err := toml.Unmarshal(data, &f); err != nil {
		return nil, err
	}

	return &f, nil
}

// WriteTomlFile writes a File to disk
func (f *File) WriteTomlFile(fpath string) error {
	data, err := toml.Marshal(f)
	if err != nil {
		return err
	}

	return os.WriteFile(fpath, data, 0o644)
}

// MigrateFromModFile migrates a gno.mod file to gnomod.toml
func MigrateFromModFile(modFile *DeprecatedModFile, dir string) error {
	f := FromDeprecatedModFile(modFile)

	// Write the new toml file
	tomlPath := filepath.Join(dir, "gnomod.toml")
	if err := f.WriteTomlFile(tomlPath); err != nil {
		return err
	}

	// Remove the old gno.mod file
	modPath := filepath.Join(dir, "gno.mod")
	return os.Remove(modPath)
}
