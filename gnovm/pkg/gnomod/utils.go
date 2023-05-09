package gnomod

import (
	"errors"
	"os"
	"path/filepath"
)

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

	return "", errors.New("gno.mod file not found in current or any parent directory")
}
