// PKGPATH: gno.land/r/launderrh_vf
package launderrh_vf

// UPDATED (view-by-default): a /p/-method write on an object another realm
// owns now borrows only if the owner granted mutable(x). None of these
// fixtures grant, so the write described below is refused; the expected
// output pins the refusal. The original analysis is kept for context.

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
// cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): i<VPBlock(1,0)>.Field
