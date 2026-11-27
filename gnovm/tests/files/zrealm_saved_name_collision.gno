// PKGPATH: gno.land/r/zrealm_saved_name_collision
// R2a attack: a stored value called through a local name that hides the
// callee must not dodge the check.

package zrealm_saved_name_collision

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.AliasTable(cross(cur))
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
