// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"chain/runtime"

	crossrealm "gno.land/r/tests/vm/crossrealm"
)

type fooer struct{}

func (fooer) Foo(cur realm) {
	println("hello " + runtime.CurrentRealm().PkgPath())
}

func (fooer) Bar() {
	println("hello " + runtime.CurrentRealm().PkgPath())
}

var f *fooer

func main(cur realm) {
	f = &fooer{}
	crossrealm.SetFooer(cross2(cur), f)
	// crossrealm.CallFooerFooCur(cross) <-- panics: cannot cur-call to
	// external realm function gno.land/r/crossrealm_test.Foo from
	// gno.land/r/tests/vm/crossrealm
	crossrealm.CallFooerFooCross(cross2(cur))
	crossrealm.CallFooerBar()
	crossrealm.CallFooerBarCrossing(cross2(cur))
	print(".")
}

// Output:
// hello gno.land/r/crossrealm
// hello gno.land/r/crossrealm
// hello gno.land/r/tests/vm/crossrealm
// .
