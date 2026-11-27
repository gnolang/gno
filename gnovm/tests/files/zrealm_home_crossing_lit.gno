// PKGPATH: gno.land/r/zrealm_home_crossing_lit
package zrealm_home_crossing_lit

// Crossing closures are never stripped: a composite literal of cbvictim's
// crossing closure (the grc20reg shape) keeps a reference and still writes.

import (
	"gno.land/r/tests/vm/cbother"
	"gno.land/r/tests/vm/cbvictim"
)

type handle struct{ set func(cur realm, s string) }

var hs []*handle

func main(cur realm) {
	hs = append(hs, &handle{set: cbvictim.SetX})
	hs[0].set(cross(cur), "lit")
	println("owner:", cbvictim.Owner)
	cbother.Register(cross(cur), cbvictim.SetX)
	cbother.CallAll(cross(cur))
	println("owner:", cbvictim.Owner)
}

// Output:
// owner: lit
// owner: crossed

// Storage:
// gno.land/r/tests/vm/cbother: 1278
// gno.land/r/tests/vm/cbvictim: 23
// gno.land/r/zrealm_home_crossing_lit: 1285
