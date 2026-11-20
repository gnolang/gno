// PKGPATH: gno.land/r/launderrh_mv_attacker_stored_mv
package launderrh_mv_attacker_stored_mv

// Explicit-entry rule: a /p/ method called from here has no power over
// another realm's object, so the write below is refused. The analysis below
// was written before the explicit-entry rule.

// Bound-method-value stored in /r/-ATTACKER state and invoked from
// /r/-attacker context. Question: does the recv-borrow borrow rule #2 still
// fire at invocation, or is it gated on the call site being /r/-victim?
//
// 1. Attacker reads &we.Immutable (aliased to victim memory).
// 2. Attacker BINDS captured.BumpToPwn — recv-bound method value.
// 3. Attacker STORES it as an /r/-attacker package variable.
//    (Note: storing involves a heap-alloc into /r/-attacker's state.)
// 4. Attacker invokes the stored mv() from main(). m.Realm =
//    /r/-attacker; PushFrameCall for BumpToPwn sees the recv stamp
//    is /r/-victim → borrow rule #2 → m.Realm = /r/-victim → BumpToPwn body
//    writes → SUCCEEDS.
//
// Direct from main this is the same as the inline call we.Immutable.
// BumpToPwn() variant of the exploit (B-class). The storage layer
// adds no safety. The point of this filetest is to confirm the
// borrow rule #2 receiver-borrow fires at INVOCATION time, regardless of
// where the function value was constructed/persisted.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

var stored func()

func main(cur realm) {
	println("before:", launderrvictim.ReadEmbed())

	we := launderrvictim.GetWithEmbed()
	captured := &we.Immutable
	stored = captured.BumpToPwn // bound mv kept in /r/-attacker state
	stored()

	println("after: ", launderrvictim.ReadEmbed())
}

// Output:
// before: embed-orig

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field
