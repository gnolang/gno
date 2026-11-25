// PKGPATH: gno.land/r/zrealm_home_own_table
package zrealm_home_own_table

// cbvictim's gate runs its own table: R2a still refuses, the values were
// made in another crossing call.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunOwnTable(cross(cur))
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
