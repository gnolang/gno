// PKGPATH: gno.land/r/zrealm_saved_returned_reentry
// R2a attack: a fresh value returned to the caller and passed back in a
// second crossing call must not write.

package zrealm_saved_returned_reentry

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	f := cbsaved.MakeValue(cross(cur))
	cbsaved.CallValue(cross(cur), f)
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/cbsaved) package{})).Owner
