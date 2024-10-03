// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/r/demo/tests"
	"std"
)

func main() {
	tests.InitTestNodes()
	if std.CurrentRealm().PkgPath() != "gno.land/r/test" {
		panic("should not happen")
	}
	tests.ModTestNodes()
	if std.CurrentRealm().PkgPath() != "gno.land/r/test" {
		panic("should not happen")
	}
	std.ClearStoreCache()
	if std.CurrentRealm().PkgPath() != "gno.land/r/test" {
		panic("should not happen")
	}
	tests.PrintTestNodes()
	if std.CurrentRealm().PkgPath() != "gno.land/r/test" {
		panic("should not happen")
	}
}

// Output:
// second's child
