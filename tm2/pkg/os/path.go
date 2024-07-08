package os

import (
	"path/filepath"
)

func MakeAbs(path string, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
