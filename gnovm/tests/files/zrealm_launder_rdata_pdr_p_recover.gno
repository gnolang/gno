// PKGPATH: gno.land/r/launderrh_pdr_p_recover
package launderrh_pdr_p_recover

// Explicit-entry rule: a function literal written here has no power over
// another realm's object, so the write below is refused. The analysis below
// was written before the explicit-entry rule.

// Vector 8: /p/-method RecoverApply has its own defer/recover.
// What happens when an attacker /r/-closure panics from inside?
//
// Trace:
//  1. main → m.Realm = attacker
//  2. inner.RecoverApply(closure) → borrow rule #2 receiver-borrow on
//     *launderpkg.Immutable (victim-stamped) → m.Realm = victim
//  3. RecoverApply schedules defer { recover() } in /p/-frame
//  4. RecoverApply calls fn(i): fn is /r/-attacker-declared closure
//     → borrow rule #1 → m.Realm = attacker
//  5. fn body: i.Field = ... → readonly check fires (storage is
//     victim, m.Realm = attacker)
//  6. Unwind fn's frame: LastRealm = victim, crlm = attacker →
//     isRealmBoundary = true at fn's frame → abort transaction
//
// The recover() in RecoverApply never runs. Cross-realm panic
// aborts before reaching it.

import (
	"gno.land/p/demo/tests/launderpkg/v0"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadPtr())

	inner := launderrvictim.GetWithPtr().Inner
	rec := inner.RecoverApply(func(i *launderpkg.Immutable) {
		i.Field = "pwnd-via-p-recover"
	})
	println("rec:", rec)

	println("after: ", launderrvictim.ReadPtr())
}

// Output:
// before: ptr-orig
// rec: cannot directly modify readonly tainted object (write it in a named function of the realm that owns it, during a crossing call into that realm; a /p/ or standard-library package's object cannot be written after its init): i<VPBlock(1,0)>.Field
// after:  ptr-orig
