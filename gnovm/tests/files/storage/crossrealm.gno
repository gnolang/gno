// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/demo/tests/crossrealm"
	"gno.land/r/demo/tests/crossrealm_b"
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
// gno.land/r/demo/tests/crossrealm: 228
// gno.land/r/demo/tests/crossrealm_b: 8

// Output:
// hello A cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// hello ABCD cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// .
