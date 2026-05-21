// PKGPATH: gno.land/r/launderrh_pdr_multilevel
package launderrh_pdr_multilevel

// Vector 10: multi-level chain — /r/-victim calls /p/-DeferApply
// which schedules attacker callback as a defer.
//
// Trace at deferred callback time:
//  1. main → m.Realm = attacker
//  2. CallPDeferApply(EvilWrite) → Rule 1 → m.Realm = victim
//  3. body: gWithPtr.Inner.DeferApply(fn) → receiver-borrow on
//     *launderpkg.Immutable victim-stamped — wait, m.Realm is
//     ALREADY victim, so the borrow check sees recvOID.PkgID ==
//     m.Realm.ID → no-op. m.Realm stays victim.
//  4. DeferApply body: schedule fn(i) as defer
//  5. DeferApply returns → defer fires → PushFrameCall(fn, true)
//     with LastRealm = victim (current m.Realm)
//  6. fn is top-level /p/ EvilWrite — no Rule 1, no Rule 2 → m.Realm
//     stays victim
//  7. EvilWrite body: i.Field = ... writes /r/-victim data under
//     m.Realm = victim → readonly passes → WRITE SUCCEEDS
//
// Now CallPDeferApply returns (defer queue is empty). The /p/-frame
// is popped, restoring m.Realm to victim (unchanged). The /r/-victim
// frame is popped, restoring m.Realm to attacker.
//
// Expected: SUCCEEDS — same as the known-open Apply pattern,
// reached through a multi-level defer chain. The defer in /p/ is
// invoked from /r/-victim-borrowed scope, so m.Realm at callback
// entry is /r/-victim.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadPtr())

	launderrvictim.CallPDeferApply(launderattack.EvilWrite)

	println("after: ", launderrvictim.ReadPtr())
}

// Output:
// before: ptr-orig
// after:  pwnd-via-apply
