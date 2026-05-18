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
	crossrealm.CallFooerFooCur(cross2(cur))
}

func main(cur realm) {
}

// Error:
// cannot resolve foreign realm RID0C8307F7BBE97FDB07B2D0FA6907B8D67517C02A for cross-realm finalize
