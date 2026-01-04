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
	f.s = "Foo"
	println("hello " + f.s + " " + runtime.CurrentRealm().PkgPath())
}

func (f *fooer) Bar() {
	f.s = "Bar"
	println("hello " + f.s + " " + runtime.CurrentRealm().PkgPath())
}

var f *fooer

func init() {
	f = &fooer{s: "A"}
	fg := func() crossrealm.Fooer { return f }
	crossrealm.SetFooerGetter(cross, fg)
	f.s = "B"
	crossrealm.CallFooerGetterFooCross(cross)
	f.s = "C"
	crossrealm.CallFooerGetterFooCross(cross)
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
