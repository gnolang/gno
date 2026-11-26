// PKGPATH: gno.land/r/launderrh_pdr_p_defer_apply
package launderrh_pdr_p_defer_apply

// Refused: a /p/ method has no power over another realm's object (R1).

// Vector 7: /p/-method DeferApply schedules attacker callback as
// a defer rather than synchronous call.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadPtr())

	inner := launderrvictim.GetWithPtr().Inner
	inner.DeferApply(launderattack.EvilWrite)

	println("after: ", launderrvictim.ReadPtr())
}

// Output:
// before: ptr-orig

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field
