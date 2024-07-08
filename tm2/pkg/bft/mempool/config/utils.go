package config

import "path/filepath"

// helper function to make config creation independent of root dir
func join(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
