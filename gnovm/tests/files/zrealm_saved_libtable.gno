// PKGPATH: gno.land/r/zrealm_saved_libtable
package zrealm_saved_libtable

// Library table called in place: today it acts for the caller realm.

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.LibInPlace(cross(cur))
	println("n:", cbsaved.LB.N)
}

// Output:
// n: 1
