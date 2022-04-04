// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/r/tests"
	"std"
)

func main() {
	if std.TestCurrentRealm() != "gno.land/r/test" {
		panic("should not happen")
	}
	tests.InitTestNodes()
	if std.TestCurrentRealm() != "gno.land/r/test" {
		panic("should not happen")
	}
	tests.ModTestNodes()
	if std.TestCurrentRealm() != "gno.land/r/test" {
		panic("should not happen")
	}
	std.ClearStoreCache()
	if std.TestCurrentRealm() != "gno.land/r/test" {
		panic("should not happen")
	}
	tests.PrintTestNodes()
	if std.TestCurrentRealm() != "gno.land/r/test" {
		panic("should not happen")
	}
}

// Output:
// second's child
