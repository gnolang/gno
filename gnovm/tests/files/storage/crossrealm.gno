// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests/crossrealm"
	"gno.land/r/demo/tests/crossrealm_b"
)

func main() {
	crossing()
	f := crossrealm_b.Fooer
	cross(crossrealm.SetFooer)(f)
	cross(crossrealm.CallFooerFooSR2)()
	f.SetS("ABCD")
	cross(crossrealm.CallFooerFooSR2)()
	print(".")
}




// Storage:
// gno.land/r/crossrealm_test: 0
// gno.land/r/demo/tests/crossrealm: 226
// gno.land/r/demo/tests/crossrealm_b: 8

// Output:
// hello A cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// hello ABCD cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// .
