// PKGPATH: gno.land/r/zrealm_home_own_helper
package zrealm_home_own_helper

// cbvictim's gate hands its own value to its own helper: R2a refuses,
// the helper is a crossing call the value did not make.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunOwnHelper(cross(cur))
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
