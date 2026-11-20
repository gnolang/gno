// PKGPATH: gno.land/r/zrealm_saved_methodvalue
package zrealm_saved_methodvalue

// S candidate: a method value of a receiver this realm owns, stored in a
// package struct and called in the same transaction.

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.BindThenRun(cross(cur))
	println("box:", cbsaved.Box.Name)
}

// Output:
// box: bound
