// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	"gno.land/r/demo/tests/crossrealm"
	"gno.land/r/demo/tests/crossrealm_b"
)

func main() {
	f := crossrealm_b.Fooer
	crossrealm.SetFooerGetter(func() crossrealm.Fooer { return f })
	crossrealm.CallFooerGetterFoo()
	f.SetS("B")
	crossrealm.CallFooerGetterFoo()
	println(".")

	f.SetS("C")
	crossrealm.SetFooerGetter(crossrealm_b.FooerGetter)
	crossrealm.CallFooerGetterFoo()
	println(".")

	f.SetS("D")
	crossrealm.SetFooerGetter(crossrealm_b.FooerGetterBuilder())
	crossrealm.CallFooerGetterFoo()
	println(".")
}

// Output:
// hello A cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// hello B cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// .
// hello C cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// .
// hello D cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// .

// Realm:

// Error:
