// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm"
	"gno.land/r/tests/vm/crossrealm_b"
)

func main(cur realm) {
	f := crossrealm_b.Fooer
	crossrealm.SetFooer(cross(cur), f)
	crossrealm.CallFooerFooCross(cross(cur))
	f.SetS("ABCD")
	crossrealm.CallFooerFooCross(cross(cur))
	print(".")
}

// Storage:
// gno.land/r/crossrealm: 0
// gno.land/r/tests/vm/crossrealm: 214
// gno.land/r/tests/vm/crossrealm_b: 25

// Output:
// hello A cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// hello ABCD cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// .
