// PKGPATH: gno.land/r/zrealm_saved_register
package zrealm_saved_register

// S candidate: a fresh closure appended to a package slice and called from a
// range copy; the value is saved, but nowhere near finalize yet.

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.RegisterThenRun(cross(cur))
	println("owner:", cbsaved.Owner)
}

// Output:
// owner: registered
