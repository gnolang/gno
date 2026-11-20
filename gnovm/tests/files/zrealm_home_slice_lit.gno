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
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
