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

var f crossrealm.Fooer

func init(cur realm) {
	println("init")
	f = crossrealm.SetFooer(cross2(cur), &fooer{})
	// crossrealm.CallFooerFooCur(cross2(cur)) <-- should panic
	crossrealm.CallFooerFooCross(cross2(cur))
	crossrealm.CallFooerBar()
	crossrealm.CallFooerBarCrossing(cross2(cur))
}

func main(cur realm) {
	println("main")
	// crossrealm.CallFooerFooCur(cross) <-- should panic
	crossrealm.CallFooerFooCross(cross2(cur))
	crossrealm.CallFooerBar()
	crossrealm.CallFooerBarCrossing(cross2(cur))
	print(".")
}

// Output:
// init
// hello gno.land/r/crossrealm
// hello gno.land/r/crossrealm
// hello gno.land/r/tests/vm/crossrealm
// main
// hello gno.land/r/crossrealm
// hello gno.land/r/crossrealm
// hello gno.land/r/tests/vm/crossrealm
// .
