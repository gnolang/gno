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
	f = &fooer{}
	c := &container{f}
	crossrealm.SetFooer(c)
	crossrealm.CallFooerFoo()
	print(".")
}

// Error:
// new escaped mark has no object ID
