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
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): Owner<~VPBlock(3,0)>
