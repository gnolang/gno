// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	"gno.land/r/demo/tests/crossrealm"
	"gno.land/r/demo/tests/crossrealm_b"
)

type Foo struct{}

func main() {
	f := crossrealm_b.Fooer
	crossrealm.SetFooerGetter(func() crossrealm.Fooer {
		f1 := Foo{}
		return f
	})
	crossrealm.CallFooerGetterFoo()
	println("ok")
}

// Error:
// cannot attach function contains object from external realm
