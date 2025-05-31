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

func main(cur realm) {
	f = &fooer{}
	c := &container{f}
	crossrealm.SetFooer(cross, c)
	crossrealm.CallFooerFoo()
	crossrealm.CallFooerFooSR(cross)
	print(".")
}

// Output:
// hello container gno.land/r/crossrealm_test
// hello container gno.land/r/demo/tests/crossrealm
// .
