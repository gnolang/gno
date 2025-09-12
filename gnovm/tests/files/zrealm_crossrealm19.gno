// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"std"

	crossrealm "gno.land/r/tests/vm/crossrealm"
)

type fooer struct {
	s string
}

func (f *fooer) Foo(cur realm) {
	// f.s = "Foo" <-- panics (crossing yet stored in crossrealm)
	println("hello " + f.s + " " + std.CurrentRealm().PkgPath())
}

func (f *fooer) Bar() {
	f.s = "Bar"
	println("hello " + f.s + " " + std.CurrentRealm().PkgPath())
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

// Error:
// invalid call: fooer<~VPBlock(3,5)>.Bar(), method realm: RIDDC8307F7BBE97FDB07B2D0FA6907B8D67517C02A does not match receiver realm: RIDA7E973C6AEAB1F1CB747D0903CB1E4D369226376
