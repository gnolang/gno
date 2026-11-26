// PKGPATH: gno.land/r/zrealm_saved_same_tx
package zrealm_saved_same_tx

// A function value stored in a package variable is read-only from the store
// on, also in the call that made it.

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.SaveThenRun(cross(cur))
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/cbsaved) package{})).Owner
