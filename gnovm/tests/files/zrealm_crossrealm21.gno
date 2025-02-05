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

var f fooer

func main() {
	f = fooer{name: "local_fooer"}
	crossrealm.SetContainer2(&f) // the object attached is package block, escaped, so this works.
	crossrealm.RunF()
	print(".")
}

// Output:
// hello gno.land/r/demo/tests/crossrealm/iface
// .
