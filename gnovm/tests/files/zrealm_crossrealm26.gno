// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f fooer

// Pointers to external realm objects that are not yet
// attached to the external realm may not be associated
// with an attached object.

func main() {
	crossrealm.AttachContainer() // this attach container first
	f = fooer{}                  // not attached to local realm yet, cannot be able to attach to external attached even by reference
	crossrealm.SetContainer(&f)
	print(".")
}

// Output:
// .

// Error:

// Realm:
