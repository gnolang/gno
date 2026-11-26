// PKGPATH: gno.land/r/tests/vm/lendplibinit
package lendplibinit

import plib "gno.land/p/demo/tests/lendcalllib"

// A library's init writes its own state through a function value: its
// realm is frozen after init, so R2a's read-only switch does not apply.
func main(cur realm) {
	println("p01 library init wrote", plib.N())
}

// Output:
// p01 library init wrote 2
