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

var f *fooer

func main(cur realm) {
	f = &fooer{}
	crossrealm.SetFooer(cross(cur), f)
	crossrealm.CallFooerFooCur(cross(cur))
}

// Error:
// cannot cur-call to external realm function gno.land/r/crossrealm.Foo from gno.land/r/tests/vm/crossrealm
