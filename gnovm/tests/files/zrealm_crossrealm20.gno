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

func (f *fooer) Bar() {
	f.s = "Bar"
	println("hello " + f.s + " " + unsafe.CurrentRealm().PkgPath())
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
	crossrealm.CallFooerGetterBar()
}

func main(cur realm) {
	print(".")
}

// Output:
// hello Foo gno.land/r/crossrealm
// hello Foo gno.land/r/crossrealm
// hello Bar gno.land/r/crossrealm
// .

// Realm:
// finalizerealm["gno.land/r/crossrealm"]
