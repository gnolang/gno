// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo() { println("hello") }

var f *fooer

func init() {
	f = &fooer{} // unreal
	crossrealm.SetFooer(f)
	crossrealm.CallFooerFoo()
}

func main() {
	print(".")
}

// Error:
// new escaped mark has no object ID
