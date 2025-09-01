// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/tests/crossrealm"
	"gno.land/r/tests/tests/crossrealm_b"
)

func main() {
	f := crossrealm_b.Fooer
	crossrealm.SetFooer(cross, f)
	crossrealm.CallFooerFooCross(cross)
	f.SetS("ABCD")
	crossrealm.CallFooerFooCross(cross)
	print(".")
}

// Storage:
// gno.land/r/crossrealm: 0
// gno.land/r/tests/tests/crossrealm: 227
// gno.land/r/tests/tests/crossrealm_b: 8

// Output:
// hello A cur=gno.land/r/tests/tests/crossrealm_b prev=gno.land/r/tests/tests/crossrealm
// hello ABCD cur=gno.land/r/tests/tests/crossrealm_b prev=gno.land/r/tests/tests/crossrealm
// .
