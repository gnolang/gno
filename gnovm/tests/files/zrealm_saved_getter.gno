// PKGPATH: gno.land/r/zrealm_saved_getter
package zrealm_saved_getter

// S candidate: a value returned by this realm's own non-crossing getter.

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.CallThroughGetter(cross(cur))
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
