// PKGPATH: gno.land/r/zrealm_home_known_ptr
package zrealm_home_known_ptr

// R2a: the gate reaches cbvictim's own value through a *func pointer this
// realm supplied; the value was made in another crossing call, so it is read-only here.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunP(cross(cur), &cbvictim.FV)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
