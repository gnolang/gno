// PKGPATH: gno.land/r/launderrh_vf
package launderrh_vf

// Explicit-entry rule: a /p/ method called from here has no power over
// another realm's object, so the write below is refused. The analysis below
// was written before the explicit-entry rule.

// /r/-declared victim with /p/-VALUE FIELD — Apply-callback attack.
//
// WithVal has `Inner launderpkg.Immutable` (named, not embedded).
// Attacker takes the address of the inner field, then calls Apply.
// Same exploit pattern as the embed/pointer variants.

import (
	"gno.land/p/demo/tests/launderattack/v0"
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

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field
