// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
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
	crossrealm.CallFooerFooCur(cross)
}

func main(cur realm) {
}

// Output:

// Error:
// cannot cur-call to external realm function gno.land/r/crossrealm_test.Foo from gno.land/r/demo/tests/crossrealm
