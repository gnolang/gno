// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"chain/runtime/unsafe"

	crossrealm "gno.land/r/tests/vm/crossrealm"
)

type fooer struct{}

func (fooer) Foo(cur realm) {
	println("hello " + unsafe.CurrentRealm().PkgPath())
}

func (fooer) Bar() {
	println("hello " + unsafe.CurrentRealm().PkgPath())
}

var f crossrealm.Fooer

func init(cur realm) {
	println("init")
	f = crossrealm.SetFooer(cross(cur), &fooer{})
	// crossrealm.CallFooerFooCur(cross(cur)) <-- should panic
	crossrealm.CallFooerFooCross(cross(cur))
	crossrealm.CallFooerBar()
	crossrealm.CallFooerBarCrossing(cross(cur))
}

func main(cur realm) {
	println("main")
	// crossrealm.CallFooerFooCur(cross) <-- should panic
	crossrealm.CallFooerFooCross(cross(cur))
	crossrealm.CallFooerBar()
	crossrealm.CallFooerBarCrossing(cross(cur))
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
