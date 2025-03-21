// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct {
	s string
}

func (f *fooer) Foo() {
	f.s = "B"
	println("hello")
}

var f *fooer

func init() {
	f = &fooer{s: "A"} // unreal, refCount 1, owner f
	crossrealm.SetFooer(f)
	crossrealm.CallFooerFoo()
}

func main() {
	print(".")
}

// Error:
// cannot attach a reference to an unattached object from an external realm
