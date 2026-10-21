// PKGPATH: gno.land/r/launderrh_pdr_p_panic
package launderrh_pdr_p_panic

// UPDATED (view-by-default): a /p/-method write on an object another realm
// owns now borrows only if the owner granted mutable(x). None of these
// fixtures grant, so the write described below is refused; the expected
// output pins the refusal. The original analysis is kept for context.

// Vector 11: a /p/-helper called from attacker panics. Since no
// realm boundary is crossed during unwind (the /p/-frame's LastRealm
// = attacker), the panic is catchable by attacker's defer.
//
// Trace:
//  1. main → m.Realm = attacker
//  2. attacker.bad() (local fn) → no borrow rule #1 (not borrowed), m.Realm
//     stays attacker (local fn in attacker pkg).
//  3. inside, attacker calls /p/.SomeFn that panics deliberately
//     OR triggers a write to a victim object passed in.
//
// Actually for the laundering threat, the interesting probe is:
// /p/-method writes to victim-stamped pointer (recv-borrow → m.Realm
// = victim), panic fires (cross-realm), abort. Already tested.
//
// Here we test the converse: /p/.PanicAfterApply calls EvilWrite
// (succeeds), then panics. Does the post-write panic abort, or does
// it propagate up to attacker recover normally?

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadPtr())

	defer func() {
		if r := recover(); r != nil {
			println("recovered:", r)
		} else {
			println("no recover")
		}
		println("after recover:", launderrvictim.ReadPtr())
	}()

	inner := launderrvictim.GetWithPtr().Inner
	// inner.PanicAfterApply(EvilWrite) — receiver-borrow to victim,
	// EvilWrite succeeds, then post-apply panic fires from /p/-frame
	// (m.Realm = victim at panic time).
	inner.PanicAfterApply(launderattack.EvilWrite)

	println("after: ", launderrvictim.ReadPtr())
}

// The panic post-apply originates inside a /p/-method that was
// receiver-borrowed to victim → m.Realm at panic time = victim →
// unwinding will cross the realm boundary back to attacker
// → unhandled panic, transaction aborts.
//
// CRITICAL question: does the WRITE that succeeded inside EvilWrite
// (before the panic) persist? Transaction abort should roll back
// the realm transaction.

// Output:
// before: ptr-orig
// recovered: cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): i<VPBlock(1,0)>.Field
// after recover: ptr-orig
