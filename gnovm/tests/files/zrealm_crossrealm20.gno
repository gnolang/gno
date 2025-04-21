// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct {
	s string
}

func (f *fooer) Foo() {
	f.s = "B"
	println("hello " + f.s + " " + std.CurrentRealm().PkgPath())
}

var f *fooer

func init() {
	f = &fooer{s: "A"}
	fg := func() crossrealm.Fooer { return f }
	cross(crossrealm.SetFooerGetter)(fg)
	crossrealm.CallFooerGetterFoo()
	f.s = "C"
	crossrealm.CallFooerGetterFoo()
}

func main() {
	crossing()

	print(".")
}

// Output:
// hello B gno.land/r/crossrealm_test
// hello B gno.land/r/crossrealm_test
// .

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
