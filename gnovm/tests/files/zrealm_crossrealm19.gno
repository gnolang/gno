// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"chain/runtime"

	crossrealm "gno.land/r/tests/vm/crossrealm"
)

type fooer struct {
	s string
}

func (f *fooer) Foo(cur realm) {
	// f.s = "Foo" <-- panics (crossing yet stored in crossrealm)
	println("hello " + f.s + " " + runtime.CurrentRealm().PkgPath())
}

func (f *fooer) Bar() {
	f.s = "Bar"
	println("hello " + f.s + " " + runtime.CurrentRealm().PkgPath())
}

var f *fooer

func init() {
	f = &fooer{s: "A"}
	crossrealm.SetFooer(cross, f)
	crossrealm.CallFooerFooCross(cross)
	// crossrealm.CallFooerFooCur(cross) <-- panics: cannot cross-call to
	// external realm function gno.land/r/crossrealm_test.Foo from
	// gno.land/r/tests/vm/crossrealm
	crossrealm.CallFooerBar()
	crossrealm.CallFooerBarCrossing(cross)
}

func main(cur realm) {
	print(".")
}

// Output:
// hello A gno.land/r/crossrealm
// hello Bar gno.land/r/crossrealm
// hello Bar gno.land/r/tests/vm/crossrealm
// .
