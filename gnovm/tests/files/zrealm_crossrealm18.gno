// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f crossrealm.Fooer = crossrealm.SetFooer(&fooer{})

func init() {
	crossrealm.CallFoo()
}

func main() {
	crossrealm.CallFoo()
	print(".")
}

// Output:
// hello gno.land/r/crossrealm_test
// hello gno.land/r/crossrealm_test
// .

// Error:

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
