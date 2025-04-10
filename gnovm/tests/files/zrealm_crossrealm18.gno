// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"chain/runtime"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo() { println("hello " + runtime.CurrentRealm().PkgPath()) }

var f crossrealm.Fooer = crossrealm.SetFooer(&fooer{})

func init() {
	crossrealm.CallFooerFoo()
}

func main() {
	crossrealm.CallFooerFoo()
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
