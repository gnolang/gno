// PKGPATH: gno.land/r/zrealm_saved_getter
package zrealm_saved_getter

// S candidate: a value returned by this realm's own non-crossing getter.

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.CallThroughGetter(cross(cur))
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
