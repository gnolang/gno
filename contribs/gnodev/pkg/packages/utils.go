package packages

import (
	"path/filepath"
	"strings"
)

func isGnoFile(name string) bool {
	return filepath.Ext(name) == ".gno" && !strings.HasPrefix(name, ".")
}

func isTestFile(name string) bool {
	return strings.HasSuffix(name, "_filetest.gno") || strings.HasSuffix(name, "_test.gno")
}
