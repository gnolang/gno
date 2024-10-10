// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{ name string }

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f fooer // this already attached

func main() {
	crossrealm.AttachContainer()   // this attaches container first
	f = fooer{name: "local_fooer"} // update to an already attached object
	crossrealm.SetContainer(&f)
	print(".")
}
