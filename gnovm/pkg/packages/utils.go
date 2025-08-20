package packages

import (
	"path/filepath"
)

func StdlibDir(gnoroot string, name string) string {
	return filepath.Join(gnoroot, "gnovm", "stdlibs", filepath.FromSlash(name))
}
