package main

// Some gno packages are ports of a Go package that lives at a different path,
// or of files that Go has since relocated. Comparing them against the
// identically-named Go directory reports divergence where there is none, which
// is worse than reporting nothing: it hides the real drift in noise.
//
// These tables are gno-specific and describe the *gno* side of the comparison,
// so they are applied to whichever side is the Go tree.

// dirAliases maps a gno package directory to the Go directory it is a port of.
var dirAliases = map[string]string{
	// gno ports math/rand/v2, not the v1 API: the Source interface, the
	// generator constructors and the *N method names all differ.
	"math/rand": "math/rand/v2",
}

// extraLookupDirs lists additional Go directories to search for a file that is
// absent from the directly-corresponding one. Go 1.26 moved strconv's number
// internals to internal/strconv, leaving the public package a thin wrapper, so
// gno's atof.gno/ftoa.gno/... have no counterpart under src/strconv.
var extraLookupDirs = map[string][]string{
	"strconv": {"internal/strconv"},
}

// resolveDir returns the directory the Go side should be read from for a given
// package path.
func resolveDir(pkgPath string) string {
	if alias, ok := dirAliases[pkgPath]; ok {
		return alias
	}
	return pkgPath
}

// isAliasTarget reports whether pkgPath is the Go-side target of a dirAliases
// entry. Such a directory is already covered by the aliased gno package, so
// listing it again as "missing in gno" would be wrong.
func isAliasTarget(pkgPath string) bool {
	for _, target := range dirAliases {
		if target == pkgPath {
			return true
		}
	}
	return false
}
