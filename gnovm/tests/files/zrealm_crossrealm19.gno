// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
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
	// gno.land/r/demo/tests/crossrealm
	crossrealm.CallFooerBar()
	crossrealm.CallFooerBarCrossing(cross)
}

func main(cur realm) {
	print(".")
}

// Output:
// hello A gno.land/r/crossrealm

// Error:
// should not happen, mismatched method/receiver realm, fooer<~VPBlock(3,5)>.Bar()/(gno.land/r/crossrealm---------RID1712AC7ADCFDC8E58A67E5615E20FB312394C4DF)
