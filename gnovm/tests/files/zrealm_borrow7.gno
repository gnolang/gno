// PKGPATH: gno.land/r/borrow
package borrow

import "gno.land/p/tests/vm/p_crossrealm"

func init() {
	p_crossrealm.F.Modify2() // soft cross to receiver's storage realm
	println("=====================modification discarded===========================")
}

func main() {
	println("F after modify: ", p_crossrealm.F)
	println("Err after modify: ", p_crossrealm.Err)
}

// Error:
// zrealm_borrow7.gno:4:8: unknown import path "gno.land/p/tests/vm/p_crossrealm"
