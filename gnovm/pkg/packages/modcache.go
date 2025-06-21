package packages

import (
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

func PackageDir(importPath string) string {
	return filepath.Join(gnomod.ModCachePath(), filepath.FromSlash(importPath))
}
