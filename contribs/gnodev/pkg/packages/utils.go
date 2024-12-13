package packages

import (
	"path/filepath"
	"regexp"
	"strings"
)

func isGnoFile(name string) bool {
	return filepath.Ext(name) == ".gno" && !strings.HasPrefix(name, ".")
}

func isTestFile(name string) bool {
	return strings.HasSuffix(name, "_filetest.gno") || strings.HasSuffix(name, "_test.gno")
}

var reFileName = regexp.MustCompile(`^([a-zA-Z0-9_]*\.[a-z0-9_\.]*|LICENSE|README)$`)

func isValidPackageFile(filename string) bool {
	return reFileName.MatchString(filename)
}
