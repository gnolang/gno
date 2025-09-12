// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm"
	"gno.land/r/tests/vm/crossrealm_b"
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
// gno.land/r/tests/vm/crossrealm: 224
// gno.land/r/tests/vm/crossrealm_b: 8

// Output:
// hello A cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// hello ABCD cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// .
