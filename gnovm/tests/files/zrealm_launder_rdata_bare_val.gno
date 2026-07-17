// PKGPATH: gno.land/r/launderrh_bv
package launderrh_bv

// /r/-declared container with VALUE FIELD of methods-less /p/-type.
// Attacker tries direct field write into the inner value's field.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadBareVal())

	w := launderrvictim.GetWithBareVal()
	w.Inner.Field = "pwnd-direct-val"

	println("after: ", launderrvictim.ReadBareVal())
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): w<VPBlock(1,1)>.Inner.Field
