// PKGPATH: gno.land/r/zrealm_home_known_relay_table
package zrealm_home_known_relay_table

// A reference to the realm's own stored functions: cbother hands cbvictim's
// table back by reference, and R2a refuses the write.

import (
	"gno.land/r/tests/vm/cbother"
	"gno.land/r/tests/vm/cbvictim"
)

func main(cur realm) {
	cbother.RelayTable(cross(cur))
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
