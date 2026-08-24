// Package nestedpkg provides helpers for package-path based access control.
// It is useful for upgrade patterns relying on namespaces.
//
// SECURITY: every exported helper takes `rlm realm` and reads both
// `rlm.PkgPath()` and `rlm.Previous().PkgPath()` to make an
// authorization decision. To close Class-2 designation forgery (a
// hostile realm stashes a captured realm value and passes it back to
// spoof identity), every helper gates on `rlm.IsCurrent()` first. The
// Is* predicates return false on stale rlm (fail-closed); the Assert*
// helpers panic. See docs/resources/gno-security.md.
package nestedpkg

import "strings"

// IsCallerSubPath checks if the caller realm is located in a subfolder of the current realm.
func IsCallerSubPath(_ int, rlm realm) bool {
	if !rlm.IsCurrent() {
		return false
	}
	var (
		curPath  = rlm.PkgPath() + "/"
		prevPath = rlm.Previous().PkgPath() + "/"
	)
	return strings.HasPrefix(prevPath, curPath)
}

// AssertCallerIsSubPath panics if IsCallerSubPath returns false.
func AssertCallerIsSubPath(_ int, rlm realm) {
	if !rlm.IsCurrent() {
		panic("unauthorized: rlm is not the caller's live cur")
	}
	var (
		curPath  = rlm.PkgPath() + "/"
		prevPath = rlm.Previous().PkgPath() + "/"
	)
	if !strings.HasPrefix(prevPath, curPath) {
		panic("call restricted to nested packages. current realm is " + curPath + ", previous realm is " + prevPath)
	}
}

// IsCallerParentPath checks if the caller realm is located in a parent location of the current realm.
func IsCallerParentPath(_ int, rlm realm) bool {
	if !rlm.IsCurrent() {
		return false
	}
	var (
		curPath  = rlm.PkgPath() + "/"
		prevPath = rlm.Previous().PkgPath() + "/"
	)
	return strings.HasPrefix(curPath, prevPath)
}

// AssertCallerIsParentPath panics if IsCallerParentPath returns false.
func AssertCallerIsParentPath(_ int, rlm realm) {
	if !rlm.IsCurrent() {
		panic("unauthorized: rlm is not the caller's live cur")
	}
	var (
		curPath  = rlm.PkgPath() + "/"
		prevPath = rlm.Previous().PkgPath() + "/"
	)
	if !strings.HasPrefix(curPath, prevPath) {
		panic("call restricted to parent packages. current realm is " + curPath + ", previous realm is " + prevPath)
	}
}

// IsSameNamespace checks if the caller realm and the current realm are in the same namespace.
func IsSameNamespace(_ int, rlm realm) bool {
	if !rlm.IsCurrent() {
		return false
	}
	var (
		curNs  = nsFromPath(rlm.PkgPath()) + "/"
		prevNs = nsFromPath(rlm.Previous().PkgPath()) + "/"
	)
	return curNs == prevNs
}

// AssertIsSameNamespace panics if IsSameNamespace returns false.
func AssertIsSameNamespace(_ int, rlm realm) {
	if !rlm.IsCurrent() {
		panic("unauthorized: rlm is not the caller's live cur")
	}
	var (
		curNs  = nsFromPath(rlm.PkgPath()) + "/"
		prevNs = nsFromPath(rlm.Previous().PkgPath()) + "/"
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
