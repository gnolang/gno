// PKGPATH: gno.land/r/launderrh_pf
package launderrh_pf

// UPDATED (view-by-default): a /p/-method write on an object another realm
// owns now borrows only if the owner granted mutable(x). None of these
// fixtures grant, so the write described below is refused; the expected
// output pins the refusal. The original analysis is kept for context.

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
// cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): i<VPBlock(1,0)>.Field
