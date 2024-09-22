// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f *fooer

// XXX, this attach a heapItem to current realm
func init() {
	f = &fooer{}
}

func main() {
	// XXX, this associate a pointer to a value(already attached to local realm)
	// to external realm, no attachment happens to the external realm.
	crossrealm.SetFooer(f)
	crossrealm.CallFoo()
	print(".")
}

// Output:
// hello gno.land/r/demo/tests/crossrealm
// .

// Error:

// Realm:
