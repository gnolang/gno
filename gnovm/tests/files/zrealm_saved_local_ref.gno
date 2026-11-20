// PKGPATH: gno.land/r/zrealm_saved_local_ref
// S3 attack: a local table attached by reference and called in a later
// crossing call must not write.

package zrealm_saved_local_ref

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.MakeLocal(cross(cur))
	cbsaved.AttachLocal(cross(cur))
	cbsaved.RunReg(cross(cur))
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only in the crossing call that made it: (const (ref(gno.land/r/tests/vm/cbsaved) package{})).Owner
