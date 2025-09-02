// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"std"

	crossrealm "gno.land/r/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo(cur realm) {
	println("hello " + std.CurrentRealm().PkgPath())
}

func (fooer) Bar() {
	println("hello " + std.CurrentRealm().PkgPath())
}

var f *fooer

func init() {
	println("init")
	f = &fooer{}
	crossrealm.SetFooer(cross, f)
	// crossrealm.CallFooerFooCur(cross) <-- panics: cannot cur-call to
	// external realm function gno.land/r/crossrealm_test.Foo from
	// gno.land/r/tests/crossrealm
	crossrealm.CallFooerFooCross(cross)
	crossrealm.CallFooerBar()
	crossrealm.CallFooerBarCrossing(cross)
}

func main(cur realm) {
	print(".")
}

// Output:
// init
// hello gno.land/r/crossrealm
// hello gno.land/r/crossrealm
// hello gno.land/r/tests/crossrealm
// .
