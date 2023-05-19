package gnomod

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrGnoModNotFound is returned by [FindRootDir] when, even after traversing
// up to the root directory, a gno.mod file could not be found.
var ErrGnoModNotFound = errors.New("gno.mod file not found in current or any parent directory")

// FindRootDir determines the root directory of the project which contains the
// gno.mod file. If no gno.mod file is found, [ErrGnoModNotFound] is returned.
// The given path must be absolute.
func FindRootDir(absPath string) (string, error) {
	if !filepath.IsAbs(absPath) {
		return "", errors.New("requires absolute path")
	}

	root := filepath.VolumeName(absPath) + string(filepath.Separator)
	for absPath != root {
		modPath := filepath.Join(absPath, "gno.mod")
		_, err := os.Stat(modPath)
		if errors.Is(err, os.ErrNotExist) {
			absPath = filepath.Dir(absPath)
			continue
		}
		if err != nil {
			return "", err
		}
		return absPath, nil
	}

	return "", ErrGnoModNotFound
}
