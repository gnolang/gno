// PKGPATH: gno.land/r/zrealm_saved_cross_helper
// S3's stated cost: a realm passes a fresh closure to its own crossing
// helper; the helper's crossing call refuses the write.

package zrealm_saved_cross_helper

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.CallHelper(cross(cur))
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only in the crossing call that made it: (const (ref(gno.land/r/tests/vm/cbsaved) package{})).Owner
