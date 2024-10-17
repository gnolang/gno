// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{ name string }

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f fooer

func main() {
	f = fooer{name: "local_fooer"}
	crossrealm.SetContainer3(&f)
	print(".")
}
