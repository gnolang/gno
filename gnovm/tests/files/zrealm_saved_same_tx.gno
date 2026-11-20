// PKGPATH: gno.land/r/zrealm_saved_same_tx
package zrealm_saved_same_tx

// S candidate: a value saved in the same transaction is not real yet, so the
// SAVED predicate may miss it. Stored in a package variable and called here.

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.SaveThenRun(cross(cur))
	println("owner:", cbsaved.Owner)
}

// Output:
// owner: same-tx
