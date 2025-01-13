// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var local_fooer = &fooer{}

func init() {
	var f crossrealm.Fooer = crossrealm.SetFooer(local_fooer) // until now, local_fooer is real
	crossrealm.CallFooerFoo()
}

func main() {
	crossrealm.CallFooerFoo()
	print(".")
}

// Output:
// hello gno.land/r/demo/tests/crossrealm
// hello gno.land/r/demo/tests/crossrealm
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
