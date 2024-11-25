package packages

import (
	"path/filepath"
	"strings"
)

// PathIsLocalImport reports whether the import path is
// a local import path, like ".", "..", "./foo", or "../foo".
func PathIsLocalImport(path string) bool {
	return path == "." || path == ".." ||
		strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../")
}

// PatternIsLiteral reports whether the pattern is free of wildcards.
//
// A literal pattern must match at most one package.
func PatternIsLiteral(pattern string) bool {
	return !strings.Contains(pattern, "...")
}

// PatternIsLocal reports whether the pattern must be resolved from a specific root or
// directory, such as a filesystem path or a single module.
func PatternIsLocal(pattern string) bool {
	return PathIsLocalImport(pattern) || filepath.IsAbs(pattern)
}

// PatternIsRemote reports whether the pattern is a remote, like "gno.land/p/demo/avl" or "gno.land/r/..."
func PatternIsRemote(pattern string) bool {
	if PatternIsLocal(pattern) {
		return false
	}

	parts := strings.Split(pattern, "/")
	if len(parts) < 2 {
		return false
	}

	domain := parts[0]
	return strings.ContainsRune(domain, '.')
}
