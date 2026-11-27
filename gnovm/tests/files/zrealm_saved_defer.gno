// PKGPATH: gno.land/r/zrealm_saved_defer
package zrealm_saved_defer

// S candidate: a deferred call of a saved value.

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.DeferSaved(cross(cur))
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
