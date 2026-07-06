// PKGPATH: gno.land/r/launderrh_pdr_victim_recover
package launderrh_pdr_victim_recover

// Vector 5: /r/-victim's RecoverAndRetry calls attacker callback
// h(gImm), which panics with the readonly violation. Victim's own
// defer attempts to recover the panic.
//
// Surprising finding (verified): victim's recover() does NOT catch
// the panic. Trace:
//  1. main → m.Realm = attacker
//  2. RecoverAndRetry → borrow rule #1 → m.Realm = victim
//  3. RecoverAndRetry schedules a defer (its own /r/-victim closure)
//  4. RecoverAndRetry calls h(gImm) → borrow rule #1 → m.Realm = attacker
//     (because h is /r/-attacker-declared)
//  5. h body: i.Field = ... → readonly panic
//  6. Unwind: at h's frame, LastRealm = victim, crlm = attacker →
//     isRealmBoundary = true → PopUntilLastReviveFrame returns nil
//     → makeUnhandledPanicError abort
//
// Victim's defer NEVER RUNS, because the panic crosses a realm
// boundary before reaching it. This is correct/intentional: a
// transaction touching readonly-foreign data aborts hard.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadImm())

	recovered := launderrvictim.RecoverAndRetry(func(i *launderrvictim.Immutable) {
		i.Field = "pwnd-via-victim-recover"
	})

	// Inspect recovered exception. We expect a readonly panic, not nil.
	if recovered == nil {
		println("UNEXPECTED: nil recover")
	} else {
		println("recovered some exception (write was blocked)")
	}

	// Attacker tries again, this time direct.
	defer func() {
		if r := recover(); r != nil {
			println("retry blocked:", r != nil)
		}
	}()

	imm := launderrvictim.GetImm()
	imm.Field = "pwnd-direct-after"

	println("after: ", launderrvictim.ReadImm())
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): i<VPBlock(1,0)>.Field
