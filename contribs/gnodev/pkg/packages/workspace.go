package packages

import (
	"os"
	"path/filepath"
)

// FindWorkspace walks up from start looking for gnowork.toml or gnomod.toml.
// Returns the directory containing the first match, or "" if none found.
func FindWorkspace(start string) string {
	dir, err := filepath.Abs(start)
	if err != nil {
		return ""
	}
	for {
		if hasFile(dir, "gnowork.toml") || hasFile(dir, "gnomod.toml") {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func hasFile(dir, name string) bool {
	info, err := os.Stat(filepath.Join(dir, name))
	return err == nil && !info.IsDir()
}
