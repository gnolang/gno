// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{ name string }

func (fooer) Foo() { println("hello from local realm fooer: " + std.CurrentRealm().PkgPath()) }

var local_fooer = &fooer{name: "local_fooer"}

func init() {
	crossrealm.SetFooer(local_fooer)
	crossrealm.CallFooerFoo()
}

func main() {
	crossrealm.CallFooerFoo()
	println(".")
}

// Output:
// hello from local realm fooer: gno.land/r/demo/tests/crossrealm
// hello from local realm fooer: gno.land/r/demo/tests/crossrealm
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
