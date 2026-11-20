// PKGPATH: gno.land/r/launderrh_pf
package launderrh_pf

// Explicit-entry rule: a /p/ method called from here has no power over
// another realm's object, so the write below is refused. The analysis below
// was written before the explicit-entry rule.

// /r/-declared victim with /p/-POINTER FIELD — Apply-callback attack.
//
// WithPtr has `Inner *launderpkg.Immutable`. Attacker reads the
// pointer field, then calls Apply with a /p/-declared callback.
//
// Same trace as embed variant: Apply borrows m.Realm to victim,
// /p/ callback inherits that authority, write succeeds.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadPtr())

	wp := launderrvictim.GetWithPtr()
	wp.Inner.Apply(launderattack.EvilWrite)

	println("after: ", launderrvictim.ReadPtr())
}

// Output:
// before: ptr-orig

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field
