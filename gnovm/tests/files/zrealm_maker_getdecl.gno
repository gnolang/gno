// PKGPATH: gno.land/r/zrealm_maker_getdecl
package zrealm_maker_getdecl

// A realm's own declared function handed out by read-only code has no maker:
// the realm's gate runs it read-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	println("owner:", cbvictim.Owner)
	cbvictim.RunS(cross(cur), cbvictim.GetDecl())
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
