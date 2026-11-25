// PKGPATH: gno.land/r/zrealm_home_third
package zrealm_home_third

// A third realm relays cbvictim's value back to cbvictim's gate: its hold
// is not the maker's.

import (
	"gno.land/r/tests/vm/cbother"
	"gno.land/r/tests/vm/cbvictim"
)

func main(cur realm) {
	cbother.Relay(cross(cur))
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
