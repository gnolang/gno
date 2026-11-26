// PKGPATH: gno.land/r/launderc
package launderc

// Refused: a /p/ method has no power over another realm's object (R1).

// Attacker obtains a pointer to victim's g via the victim's
// public-by-design getter, then invokes a /p/-defined method on it.

import (
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	println("before:", laundervictim.ReadPtr())

	// Obtain pointer to victim's persisted Object.
	p := laundervictim.GetPtr()

	// The /p/ method Set runs for this realm, which does not own p: refused.
	p.Set("pwnd-via-pmethod")

	println("after: ", laundervictim.ReadPtr())
}

// Output:
// before: original-ptr

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): o<VPBlock(1,0)>.Field
