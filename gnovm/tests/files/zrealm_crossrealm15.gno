// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

var f *fooer

func init() {
	f = &fooer{}
	withswitch(crossrealm.SetFooer)(f)
	crossrealm.CallFooerFoo()
	withswitch(crossrealm.CallFooerFooSR)()
}

func main() {
	switchrealm()

	print(".")
}

// Output:
// hello gno.land/r/crossrealm_test
// hello gno.land/r/demo/tests/crossrealm
// .
