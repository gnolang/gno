// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct {
	name string
}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f fooer

// XXX, this should work when reference to not attached local object
// associated to external realm *NOT* attached object.
func main() {
	f = fooer{name: "local_fooer"} // this is not attached
	crossrealm.SetContainer2(&f)   // XXX, here should work to associate *NOT* attached object to external container
	print(".")
}
