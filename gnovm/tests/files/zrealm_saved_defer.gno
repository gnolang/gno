// PKGPATH: gno.land/r/zrealm_saved_defer
package zrealm_saved_defer

// S candidate: a deferred call of a saved value.

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.DeferSaved(cross(cur))
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
