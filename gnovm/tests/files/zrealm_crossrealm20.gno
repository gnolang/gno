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
	f.s = "Foo"
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
	fg := func() crossrealm.Fooer { return f }
	crossrealm.SetFooerGetter(cross(cur), fg)
	f.s = "B"
	crossrealm.CallFooerGetterFooCross(cross(cur))
	f.s = "C"
	crossrealm.CallFooerGetterFooCross(cross(cur))
	f.s = "D"
	println("bar:", revive(func() { crossrealm.CallFooerGetterBar() }))
}

func main(cur realm) {
	print(".")
}

// Output:
// hello Foo gno.land/r/crossrealm
// hello Foo gno.land/r/crossrealm
// hello D gno.land/r/crossrealm
// bar: gno.land/r/crossrealm's code was entered without crossing and is read-only (call its crossing function): f<VPBlock(1,0)>.s
// .

// Realm:
// finalizerealm["gno.land/r/crossrealm"]
