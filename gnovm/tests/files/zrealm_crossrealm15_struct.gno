// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo() { println("hello") }

var f *fooer

func init() {
	f = &fooer{}
	crossrealm.SetFooer(f)
	crossrealm.CallFooerFoo()
}

func main() {
	print(".")
}

// Error:
// cannot attach a reference to an unreal object from an external realm: heapitem((struct{} gno.land/r/crossrealm_test.fooer))
