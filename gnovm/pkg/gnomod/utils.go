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
		if _, err := os.Stat(modPath); err == nil {
			return absPath, nil
		}
		absPath = filepath.Dir(absPath)
	}

	return "", errors.New("cannot guess gno.mod dir")
}
