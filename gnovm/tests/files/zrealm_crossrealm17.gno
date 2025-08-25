// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type container struct{ *fooer }

func (container) Foo(cur realm) {
	println("hello container " + std.CurrentRealm().PkgPath())
}

func (container) Bar() {
	println("hello container " + std.CurrentRealm().PkgPath())
}

type fooer struct{}

var f *fooer

func main(cur realm) {
	f = &fooer{}
	c := &container{f}
	crossrealm.SetFooer(cross, c)
	// crossrealm.CallFooerFooCur(cur) <- panics: cannot cur-call to
	// external realm function
	// gno.land/r/demo/tests/crossrealm.crossrealm<VPBlock(2,1)>.CallFooerFooCur
	// from gno.land/r/crossrealm_test
	// crossrealm.CallFooerFooCur(cross) <-- panics: cannot cross-call to
	// external realm function gno.land/r/crossrealm_test.Foo from
	// gno.land/r/demo/tests/crossrealm
	crossrealm.CallFooerFooCross(cross)
	crossrealm.CallFooerBar()
	crossrealm.CallFooerBarCrossing(cross)
	print(".")
}

// Output:
// hello container gno.land/r/crossrealm
// hello container gno.land/r/crossrealm
// hello container gno.land/r/demo/tests/crossrealm
// .

