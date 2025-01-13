// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f *fooer

func main() {
	f = &fooer{}
	crossrealm.SetFooer(f)
	crossrealm.CallFooerFoo()
	print(".")
}

// Error:
// cannot attach un-real object from external realm: struct{}
