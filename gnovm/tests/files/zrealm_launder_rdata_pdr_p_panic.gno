// PKGPATH: gno.land/r/launderrh_pdr_p_panic
package launderrh_pdr_p_panic

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
	"gno.land/p/demo/tests/launderattack"
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

// Error:
// post-apply panic
