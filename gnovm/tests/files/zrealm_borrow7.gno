// PKGPATH: gno.land/r/borrow
package borrow

import "gno.land/p/demo/tests/p_crossrealm"

func init() {
	p_crossrealm.F.Modify2() // soft cross to receiver's storage realm
	println("=====================modification discarded===========================")
}

func main() {
	println("F after modify: ", p_crossrealm.F)
	println("Err after modify: ", p_crossrealm.Err)
}

// Output:
// &(struct{(1 int)} gno.land/p/demo/tests/p_crossrealm.Foo)
// Err 2
// =====================modification discarded===========================
// F after modify:  readonly(&(struct{(0 int)} gno.land/p/demo/tests/p_crossrealm.Foo))
// Err after modify:  Err 1
