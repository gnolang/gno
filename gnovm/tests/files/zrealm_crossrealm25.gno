// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{ name string }

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f fooer

func init() {
	f = fooer{name: "local fooer"}
}

func main() {
	crossrealm.SetContainer(f)
	print(".")
}

// Error:
// should not happen while attempting to attach objects by value from external realm
