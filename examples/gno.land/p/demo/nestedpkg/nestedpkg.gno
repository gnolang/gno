// Package nestedpkg provides helpers for package-path based access control.
// It can be convenient for some upgrade patterns relying on namespaces.
package nestedpkg

import (
	"std"
	"strings"
)

// IsCallerSubPath checks if the caller realm is located in a subfolder of the current realm.
func IsCallerSubPath() bool {
	cur := std.CurrentRealm().PkgPath()
	prev := std.PrevRealm().PkgPath()
	return strings.HasPrefix(prev, cur)
}

// AssertCallerIsSubPath panics when IsCallerSubPath is false.
func AssertCallerIsSubPath() {
	if !IsCallerSubPath() {
		panic("call restricted to nested packages")
	}
}

// IsCallerParentPath checks if the caller realm is located in a parent location of the current realm.
func IsCallerParentPath() bool {
	cur := std.CurrentRealm().PkgPath()
	prev := std.PrevRealm().PkgPath()
	return strings.HasPrefix(cur, prev)
}

func AssertCallerIsParentPath() {
	if !IsCallerParentPath() {
		panic("call restricted to parent packages")
	}
}

// XXX: consider adding IsCallerDirectlySubPath
// XXX: consider adding IsCallerDirectlyParentPath
