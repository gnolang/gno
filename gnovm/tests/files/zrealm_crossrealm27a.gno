// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f fooer

// XXX, this should work when reference to not attached local object
// associated to external realm *NOT* attached object.
// XXX, panic should happen when the attachment happen in external realm

// In this way, all references to external realm objects that are reachable
// from a realm’s package global variables must already be attached to the external realm.
func main() {
	f = fooer{}
	crossrealm.SetContainer(&f)
	print(".")
}

// Output:
// .

// Error:

// Realm:
