// Package nestedpkg provides helpers for package-path based access control.
// It can be convenient for some upgrade patterns relying on namespaces.
package nestedpkg

import (
	"std"
	"strings"
)

// IsCallerSubPath checks if the caller realm is located in a subfolder of the current realm.
func IsCallerSubPath() bool {
	var (
		cur  = std.CurrentRealm().PkgPath() + "/"
		prev = std.PrevRealm().PkgPath() + "/"
	)
	return strings.HasPrefix(prev, cur)
}

// AssertCallerIsSubPath panics when IsCallerSubPath is false.
func AssertCallerIsSubPath() {
	var (
		cur  = std.CurrentRealm().PkgPath() + "/"
		prev = std.PrevRealm().PkgPath() + "/"
	)
	if !strings.HasPrefix(prev, cur) {
		panic("call restricted to nested packages. current realm is " + cur + ", prev realm is " + prev)
	}
}

// IsCallerParentPath checks if the caller realm is located in a parent location of the current realm.
func IsCallerParentPath() bool {
	var (
		cur  = std.CurrentRealm().PkgPath() + "/"
		prev = std.PrevRealm().PkgPath() + "/"
	)
	return strings.HasPrefix(cur, prev)
}

func AssertCallerIsParentPath() {
	var (
		cur  = std.CurrentRealm().PkgPath() + "/"
		prev = std.PrevRealm().PkgPath() + "/"
	)
	if !strings.HasPrefix(cur, prev) {
		panic("call restricted to parent packages. current realm is " + cur + ", prev realm is " + prev)
	}
}

func IsSameNamespace() bool {
	var (
		cur  = nsFromPath(std.CurrentRealm().PkgPath()) + "/"
		prev = nsFromPath(std.PrevRealm().PkgPath()) + "/"
	)
	return cur == prev
}

func AssertIsSameNamespace() {
	var (
		cur  = nsFromPath(std.CurrentRealm().PkgPath()) + "/"
		prev = nsFromPath(std.PrevRealm().PkgPath()) + "/"
	)
	if cur != prev {
		panic("call restricted to packages from the same namespace. current realm is " + cur + ", prev realm is " + prev)
	}
}

func nsFromPath(pkgpath string) string {
	parts := strings.Split(pkgpath, "/")

	// specifically about gno.land, potential paths are in the form of DOMAIN/r/NAMESPACE/...
	// XXX: consider extra checks.
	// XXX: support non gno.land domains, where p/ and r/ won't be enforced.
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

// XXX: consider adding IsCallerDirectlySubPath
// XXX: consider adding IsCallerDirectlyParentPath
