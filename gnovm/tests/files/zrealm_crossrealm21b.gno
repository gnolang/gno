// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type fooer struct {
	name string
}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

func main() {
	crossrealm.SetContainer2(&fooer{name: "local_fooer"})
	print(".")
}

// Error:
// cannot attach a value of a type defined by another realm: heapitem((struct{("local_fooer" string)} gno.land/r/crossrealm_test.fooer))
