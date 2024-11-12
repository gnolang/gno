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

func main() {
	f := fooer{name: "local_fooer"} // f is "new real" cuz it's parent object is new real
	crossrealm.SetContainer2(&f)    // attach new real from external realm, panic
	print(".")
}

// Error:
// should not happen while attempting to attach new real object from external realm
