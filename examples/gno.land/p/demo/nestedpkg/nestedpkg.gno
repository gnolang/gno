// Package nestedpkg provides helpers for package-path based access control.
// It is useful for upgrade patterns relying on namespaces.
//
// SECURITY: every exported helper reads both runtime.Self().PkgPath()
// (current realm's path) and runtime.Caller().PkgPath() (calling realm's
// path) to make an authorization decision. Identity is sourced from VM
// machine ops, not a passed-in `realm` value — no forgery surface
// (omarsy-free).
package nestedpkg

import (
	"chain/runtime"
	"strings"
)

// IsCallerSubPath checks if the caller realm is located in a subfolder of the current realm.
func IsCallerSubPath() bool {
	var (
		curPath  = runtime.Self().PkgPath() + "/"
		prevPath = runtime.Caller().PkgPath() + "/"
	)
	return strings.HasPrefix(prevPath, curPath)
}

// AssertCallerIsSubPath panics if IsCallerSubPath returns false.
func AssertCallerIsSubPath() {
	var (
		curPath  = runtime.Self().PkgPath() + "/"
		prevPath = runtime.Caller().PkgPath() + "/"
	)
	if !strings.HasPrefix(prevPath, curPath) {
		panic("call restricted to nested packages. current realm is " + curPath + ", previous realm is " + prevPath)
	}
}

// IsCallerParentPath checks if the caller realm is located in a parent location of the current realm.
func IsCallerParentPath() bool {
	var (
		curPath  = runtime.Self().PkgPath() + "/"
		prevPath = runtime.Caller().PkgPath() + "/"
	)
	return strings.HasPrefix(curPath, prevPath)
}

// AssertCallerIsParentPath panics if IsCallerParentPath returns false.
func AssertCallerIsParentPath() {
	var (
		curPath  = runtime.Self().PkgPath() + "/"
		prevPath = runtime.Caller().PkgPath() + "/"
	)
	if !strings.HasPrefix(curPath, prevPath) {
		panic("call restricted to parent packages. current realm is " + curPath + ", previous realm is " + prevPath)
	}
}

// IsSameNamespace checks if the caller realm and the current realm are in the same namespace.
func IsSameNamespace() bool {
	var (
		curNs  = nsFromPath(runtime.Self().PkgPath()) + "/"
		prevNs = nsFromPath(runtime.Caller().PkgPath()) + "/"
	)
	return curNs == prevNs
}

// AssertIsSameNamespace panics if IsSameNamespace returns false.
func AssertIsSameNamespace() {
	var (
		curNs  = nsFromPath(runtime.Self().PkgPath()) + "/"
		prevNs = nsFromPath(runtime.Caller().PkgPath()) + "/"
	)
	if curNs != prevNs {
		panic("call restricted to packages from the same namespace. current realm is " + curNs + ", previous realm is " + prevNs)
	}
}

// nsFromPath extracts the namespace from a package path.
func nsFromPath(pkgpath string) string {
	parts := strings.Split(pkgpath, "/")

	// Specifically for gno.land, potential paths are in the form of DOMAIN/r/NAMESPACE/...
	// XXX: Consider extra checks.
	// XXX: Support non gno.land domains, where p/ and r/ won't be enforced.
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

// XXX: Consider adding IsCallerDirectlySubPath
// XXX: Consider adding IsCallerDirectlyParentPath
