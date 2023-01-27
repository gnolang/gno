package config

import "path/filepath"

// join path to root unless path is already absolute.
func join(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(root, path)
}
