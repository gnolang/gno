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
	crossrealm.CallFooerFooCur(cross(cur))
}

func main(cur realm) {
}

// Error:
// cannot cur-call to external realm function gno.land/r/crossrealm.Foo from gno.land/r/tests/vm/crossrealm
