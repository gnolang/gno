// PKGPATH: gno.land/r/zrealm_saved_panic_recover
package zrealm_saved_panic_recover

// S candidate: a saved value carried out through panic and run after recover.

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.PanicWithSaved(cross(cur))
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
