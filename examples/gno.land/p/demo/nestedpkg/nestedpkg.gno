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
	cur := std.CurrentRealm().PkgPath()
	prev := std.PrevRealm().PkgPath()
	if !strings.HasPrefix(prev, cur) {
		panic("call restricted to nested packages. current realm is " + cur + ", prev realm is " + prev)
	}
}

// IsCallerParentPath checks if the caller realm is located in a parent location of the current realm.
func IsCallerParentPath() bool {
	cur := std.CurrentRealm().PkgPath()
	prev := std.PrevRealm().PkgPath()
	return strings.HasPrefix(cur, prev)
}

func AssertCallerIsParentPath() {
	cur := std.CurrentRealm().PkgPath()
	prev := std.PrevRealm().PkgPath()
	if !strings.HasPrefix(cur, prev) {
		panic("call restricted to parent packages. current realm is " + cur + ", prev realm is " + prev)
	}
}

// XXX: consider adding IsCallerDirectlySubPath
// XXX: consider adding IsCallerDirectlyParentPath
// XXX: consider adding IsSameNamespace
