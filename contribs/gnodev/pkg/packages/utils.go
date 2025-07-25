package packages

import (
	"path/filepath"
	"strings"
)

func isGnoFile(name string) bool {
	return filepath.Ext(name) == ".gno" && !strings.HasPrefix(name, ".")
}
