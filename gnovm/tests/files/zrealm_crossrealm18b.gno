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
	crossrealm.CallFooerFooCur(cross)
}

func main(cur realm) {
}

// Error:
// cannot cur-call to external realm function gno.land/r/crossrealm.Foo from gno.land/r/tests/vm/crossrealm
