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

var f crossrealm.Fooer = crossrealm.SetFooer(cross, &fooer{})

func init(cur realm) {
	println("init")
	// crossrealm.CallFooerFooCur(cross) <-- should panic
	crossrealm.CallFooerFooCross(cross)
	crossrealm.CallFooerBar()
	crossrealm.CallFooerBarCrossing(cross)
}

func main(cur realm) {
	println("main")
	// crossrealm.CallFooerFooCur(cross) <-- should panic
	crossrealm.CallFooerFooCross(cross)
	crossrealm.CallFooerBar()
	crossrealm.CallFooerBarCrossing(cross)
	print(".")
}

// Output:
// init
// hello gno.land/r/crossrealm
// hello gno.land/r/crossrealm
// hello gno.land/r/tests/crossrealm
// main
// hello gno.land/r/crossrealm
// hello gno.land/r/crossrealm
// hello gno.land/r/tests/crossrealm
// .

// Realm:
// finalizerealm["gno.land/r/crossrealm"]
// finalizerealm["gno.land/r/tests/crossrealm"]
// finalizerealm["gno.land/r/tests/crossrealm"]
// finalizerealm["gno.land/r/crossrealm"]
