// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"chain/runtime"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct{}

func (fooer) Foo() { println("hello " + runtime.CurrentRealm().PkgPath()) }

var f *fooer

func main() {
	f = &fooer{}
	crossrealm.SetFooer(f)
	crossrealm.CallFooerFoo()
	print(".")
}

// Error:
// new escaped mark has no object ID
