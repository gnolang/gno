// PKGPATH: gno.land/r/zrealm_saved_register
package zrealm_saved_register

// A fresh closure appended to a package slice is read-only from the store on,
// also through a range copy in the same call.

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.RegisterThenRun(cross(cur))
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/cbsaved) package{})).Owner
