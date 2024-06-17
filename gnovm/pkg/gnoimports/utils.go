package gnoimports

import (
	"path/filepath"
	"strings"
)

func isGnoFile(name string) bool {
	return filepath.Ext(name) == ".gno" && !strings.HasPrefix(name, ".")
}

// isPublicGnoFile is the same as `isGnoFile` except that it will also ignore tests files.
func isPublicGnoFile(name string) bool {
	return isGnoFile(name) &&
		// Ignore testfile
		!strings.HasSuffix(name, "_filetest.gno") &&
		!strings.HasSuffix(name, "_test.gno")
}

// isPredeclared reports whether an identifier is predeclared.
func isPredeclared(s string) bool {
	return predeclaredTypes[s] || predeclaredFuncs[s] || predeclaredConstants[s]
}

var (
	predeclaredTypes = map[string]bool{
		"any":        true,
		"bool":       true,
		"byte":       true,
		"comparable": true,
		"complex64":  true,
		"complex128": true,
		"error":      true,
		"float32":    true,
		"float64":    true,
		"int":        true,
		"int8":       true,
		"int16":      true,
		"int32":      true,
		"int64":      true,
		"rune":       true,
		"string":     true,
		"uint":       true,
		"uint8":      true,
		"uint16":     true,
		"uint32":     true,
		"uint64":     true,
		"uintptr":    true,
	}
	predeclaredFuncs = map[string]bool{
		"append":  true,
		"cap":     true,
		"close":   true,
		"complex": true,
		"copy":    true,
		"delete":  true,
		"imag":    true,
		"len":     true,
		"make":    true,
		"new":     true,
		"panic":   true,
		"print":   true,
		"println": true,
		"real":    true,
		"recover": true,
	}
	predeclaredConstants = map[string]bool{
		"false": true,
		"iota":  true,
		"nil":   true,
		"true":  true,
	}
)
