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
	crossrealm.SetFooer(cross, f)
	crossrealm.CallFooerFoo()
	crossrealm.CallFooerFooSR(cross)
}

func main(cur realm) {
	print(".")
}

// Output:
// hello gno.land/r/crossrealm_test
// hello gno.land/r/demo/tests/crossrealm
// .
