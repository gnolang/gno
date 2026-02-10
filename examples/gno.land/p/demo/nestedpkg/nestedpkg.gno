// Package nestedpkg provides helpers for package-path based access control.
// It is useful for upgrade patterns relying on namespaces.
package nestedpkg

// To test this from a realm and have runtime.CurrentRealm/PreviousRealm work correctly,
// this file is tested from gno.land/r/tests/vm/nestedpkg_test.gno
// XXX: move test to ths directory once we support testing a package and
// specifying values for both PreviousRealm and CurrentRealm.

import (
	"chain/runtime"
	"strings"
)

// IsCallerSubPath checks if the caller realm is located in a subfolder of the current realm.
func IsCallerSubPath() bool {
	var (
		cur  = runtime.CurrentRealm().PkgPath() + "/"
		prev = runtime.PreviousRealm().PkgPath() + "/"
	)
	return strings.HasPrefix(prev, cur)
}

// AssertCallerIsSubPath panics if IsCallerSubPath returns false.
func AssertCallerIsSubPath() {
	var (
		cur  = runtime.CurrentRealm().PkgPath() + "/"
		prev = runtime.PreviousRealm().PkgPath() + "/"
	)
	if !strings.HasPrefix(prev, cur) {
		panic("call restricted to nested packages. current realm is " + cur + ", previous realm is " + prev)
	}
}

// IsCallerParentPath checks if the caller realm is located in a parent location of the current realm.
func IsCallerParentPath() bool {
	var (
		cur  = runtime.CurrentRealm().PkgPath() + "/"
		prev = runtime.PreviousRealm().PkgPath() + "/"
	)
	return strings.HasPrefix(cur, prev)
}

// AssertCallerIsParentPath panics if IsCallerParentPath returns false.
func AssertCallerIsParentPath() {
	var (
		cur  = runtime.CurrentRealm().PkgPath() + "/"
		prev = runtime.PreviousRealm().PkgPath() + "/"
	)
	if !strings.HasPrefix(cur, prev) {
		panic("call restricted to parent packages. current realm is " + cur + ", previous realm is " + prev)
	}
}

// IsSameNamespace checks if the caller realm and the current realm are in the same namespace.
func IsSameNamespace() bool {
	var (
		cur  = nsFromPath(runtime.CurrentRealm().PkgPath()) + "/"
		prev = nsFromPath(runtime.PreviousRealm().PkgPath()) + "/"
	)
	return cur == prev
}

// AssertIsSameNamespace panics if IsSameNamespace returns false.
func AssertIsSameNamespace() {
	var (
		cur  = nsFromPath(runtime.CurrentRealm().PkgPath()) + "/"
		prev = nsFromPath(runtime.PreviousRealm().PkgPath()) + "/"
	)
	if cur != prev {
		panic("call restricted to packages from the same namespace. current realm is " + cur + ", previous realm is " + prev)
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
