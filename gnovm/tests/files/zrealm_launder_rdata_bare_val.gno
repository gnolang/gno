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
// cannot directly modify readonly tainted object (write it in a named function of the realm that owns it, during a crossing call into that realm; a /p/ or standard-library package's object cannot be written after its init): w<VPBlock(1,1)>.Inner.Field
