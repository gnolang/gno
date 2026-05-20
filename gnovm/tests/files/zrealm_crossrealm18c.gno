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
	crossrealm.CallFooerFooCur(cross)
}

// Error:
// cannot resolve foreign realm RID0C8307F7BBE97FDB07B2D0FA6907B8D67517C02A for cross-realm finalize
