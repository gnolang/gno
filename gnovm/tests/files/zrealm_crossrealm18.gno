// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f crossrealm.Fooer = cross(crossrealm.SetFooer)(&fooer{})

func init() {
	crossrealm.CallFooerFoo()
}

func main() {
	crossing()

	crossrealm.CallFooerFoo()
	print(".")
}

// Output:
// hello gno.land/r/crossrealm_test
// hello gno.land/r/crossrealm_test
// .

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
