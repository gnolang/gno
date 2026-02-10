package gnomod

import (
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
)

// ModCachePath returns the path for gno modules
func ModCachePath() string {
	return filepath.Join(gnoenv.HomeDir(), "pkg", "mod")
}

// IsGnomodRoot returns true if the given directory contains a gnomod.toml or gno.mod file.
func IsGnomodRoot(dir string) bool {
	for _, fname := range []string{"gnomod.toml", "gno.mod"} {
		if _, err := os.Stat(filepath.Join(dir, fname)); err == nil {
			return true
		}
	}
	return false
}
