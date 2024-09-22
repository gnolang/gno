// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type container struct{ *fooer }

func (container) Foo() { println("hello container " + std.CurrentRealm().PkgPath()) }

type fooer struct{}

var f *fooer

func main() {
	// not finalize yet, so this reference can not be attached to external realm,
	// can only be associated to un-attached objects, only until the un-attached
	// objects has not been attached to the external realm yet.
	f = &fooer{}
	c := &container{f}
	crossrealm.SetFooer(c)
	crossrealm.CallFoo()
	print(".")
}

// Output:
// hello container gno.land/r/demo/tests/crossrealm
// .
