// PKGPATH: gno.land/r/zrealm_saved_cross_helper
// R2a's stated cost: a realm passes a fresh closure to its own crossing
// helper; the helper's crossing call refuses the write.

package zrealm_saved_cross_helper

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.CallHelper(cross(cur))
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/cbsaved) package{})).Owner
