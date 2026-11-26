// PKGPATH: gno.land/r/zrealm_home_slice_lit
package zrealm_home_slice_lit

// A slice literal copies its elements under this realm's hold, so the gate
// iterating the caller's slice finds no maker.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunAll(cross(cur), []func(string){cbvictim.FV})
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
