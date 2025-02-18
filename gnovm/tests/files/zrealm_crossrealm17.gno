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
	c := &container{f} // heapItemValue not real, so trying attach by value
	crossrealm.SetFooer(c)
	crossrealm.CallFooerFoo()
	print(".")
}

// Error:
// cannot attach a value of a type defined by another realm
