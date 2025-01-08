package packages

import (
	"strings"
)

// patternIsRemote reports wether a pattern is starting with a domain
func patternIsRemote(path string) bool {
	if len(path) == 0 {
		return false
	}
	if path[0] == '.' {
		return false
	}
	slashIdx := strings.IndexRune(path, '/')
	if slashIdx == -1 {
		return false
	}
	return strings.ContainsRune(path[:slashIdx], '.')
}

// patternIsLiteral reports whether the pattern is free of wildcards.
//
// A literal pattern must match at most one package.
func patternIsLiteral(pattern string) bool {
	return !strings.Contains(pattern, "...")
}
