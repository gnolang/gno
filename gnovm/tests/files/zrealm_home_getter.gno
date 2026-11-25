// PKGPATH: gno.land/r/zrealm_home_getter
package zrealm_home_getter

// A getter run read-only returns cbvictim's own value; this realm handing it
// to cbvictim's gate strips the maker.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.GetFV())
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
