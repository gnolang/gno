// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"chain/runtime/unsafe"

	crossrealm "gno.land/r/tests/vm/crossrealm"
)

type fooer struct {
	s string
}

func (f *fooer) Foo(cur realm) {
	// f.s = "Foo" would commit: Foo is a crossing method (see crossrealm20).
	println("hello " + f.s + " " + unsafe.CurrentRealm().PkgPath())
}

// Bar is non-crossing: called by another realm it runs read-only, so it
// prints and then its write to the stored f is refused.
func (f *fooer) Bar() {
	println("hello " + f.s + " " + unsafe.CurrentRealm().PkgPath())
	f.s = "Bar"
}

var f *fooer

func init(cur realm) {
	f = &fooer{s: "A"}
	crossrealm.SetFooer(cross(cur), f)
	crossrealm.CallFooerFooCross(cross(cur))
	// crossrealm.CallFooerFooCur(cross) <-- panics: cannot cross-call to
	// external realm function gno.land/r/crossrealm_test.Foo from
	// gno.land/r/tests/vm/crossrealm
	println("bar:", revive(func() { crossrealm.CallFooerBar() }))
	println("bar crossing:", revive(func() { crossrealm.CallFooerBarCrossing(cross(cur)) }))
}

func main(cur realm) {
	print(".")
}

// Output:
// hello A gno.land/r/crossrealm
// hello A gno.land/r/crossrealm
// bar: gno.land/r/crossrealm's code was entered without crossing and is read-only (call its crossing function): f<VPBlock(1,0)>.s
// hello A gno.land/r/tests/vm/crossrealm
// bar crossing: gno.land/r/crossrealm's code was entered without crossing and is read-only (call its crossing function): f<VPBlock(1,0)>.s
// .
