// PKGPATH: gno.land/r/zrealm_saved_name_collision
// S3 attack: a stored value called through a local name that hides the
// callee must not dodge the check.

package zrealm_saved_name_collision

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.AliasTable(cross(cur))
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
