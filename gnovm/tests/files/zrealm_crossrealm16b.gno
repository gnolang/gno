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
	crossrealm.SetFooer(cross, f)
	crossrealm.CallFooerFooCur(cross)
}

// Error:
// cannot cur-call to external realm function gno.land/r/crossrealm.Foo from gno.land/r/tests/vm/crossrealm
