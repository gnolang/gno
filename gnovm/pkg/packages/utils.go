package packages

import (
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
)

func StdlibDir(name string) string {
	root := gnoenv.RootDir()
	return filepath.Join(root, "gnovm", "stdlibs", filepath.FromSlash(name))
}
