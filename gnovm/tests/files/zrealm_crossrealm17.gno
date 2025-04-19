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
	crossing()

	f = &fooer{}
	c := &container{f}
	cross(crossrealm.SetFooer)(c)
	crossrealm.CallFooerFoo()
	cross(crossrealm.CallFooerFooSR)()
	print(".")
}

// Output:
// hello container gno.land/r/crossrealm_test
// hello container gno.land/r/demo/tests/crossrealm
// .
