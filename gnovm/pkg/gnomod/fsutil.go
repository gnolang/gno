package gnomod

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
)

// ModCachePath returns the path for gno modules
func ModCachePath() string {
	return filepath.Join(gnoenv.HomeDir(), "pkg", "mod")
}

// FindRootDir determines the root directory of the project which contains the
// gno.mod file. If no gnomod.toml or gno.mod file is found, [ErrNoModFile]
// is returned.
// The given path must be absolute.
func FindRootDir(absPath string) (string, error) {
	absPath = filepath.Clean(absPath)
	if !filepath.IsAbs(absPath) {
		return "", errors.New("requires absolute path")
	}

	// mount point for absPath (e.g., "/" on Unix or "C:\" on Windows)
	mountPoint := filepath.VolumeName(absPath) + string(filepath.Separator)

	curPath := absPath
	// Check if we're still within the mount point
	for strings.HasPrefix(curPath, mountPoint) && curPath != mountPoint {
		for _, fname := range []string{"gnomod.toml", "gno.mod"} {
			modPath := filepath.Join(curPath, fname)
			_, err := os.Stat(modPath)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return "", err
			}
			return curPath, nil
		}

		// curPath = curPath/..
		parent := filepath.Dir(curPath)
		if parent == curPath {
			break
		}
		curPath = parent
	}

	return "", ErrNoModFile
}
