// PKGPATH: gno.land/r/zrealm_saved_local_ref
// R2a attack: a local table attached by reference and called in a later
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
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/cbsaved) package{})).Owner
