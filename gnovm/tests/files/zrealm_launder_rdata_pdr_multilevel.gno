// PKGPATH: gno.land/r/launderrh_pdr_multilevel
package launderrh_pdr_multilevel

// Vector 10: the victim's /p/ DeferApply defers the attacker's top-level
// /p/ EvilWrite, a value, so EvilWrite runs view-only and its write is
// refused even though the victim's realm is current.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadPtr())

	launderrvictim.CallPDeferApply(launderattack.EvilWrite)

	println("after: ", launderrvictim.ReadPtr())
}

// Output:
// before: ptr-orig

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field
