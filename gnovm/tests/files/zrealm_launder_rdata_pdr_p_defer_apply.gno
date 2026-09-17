// PKGPATH: gno.land/r/launderrh_pdr_p_defer_apply
package launderrh_pdr_p_defer_apply

// UPDATED (view-by-default): a /p/-method write on an object another realm
// owns now borrows only if the owner granted mutable(x). None of these
// fixtures grant, so the write described below is refused; the expected
// output pins the refusal. The original analysis is kept for context.

// Vector 7: /p/-method DeferApply schedules attacker callback as
// a defer rather than synchronous call.
//
// Trace:
//   1. main → m.Realm = /r/launderrh_pdr_p_defer_apply
//   2. inner := victim.GetWithPtr().Inner — read access fine
//   3. inner.DeferApply(EvilWrite) — borrow rule #2 receiver-borrow on
//      *launderpkg.Immutable (victim-stamped) → m.Realm = victim
//   4. DeferApply body: schedule fn(i) as defer
//   5. DeferApply returns → defer fires → PushFrameCall(fn, true)
//      with LastRealm = victim (current m.Realm)
//   6. EvilWrite is top-level /p/ fn → no borrow rule #1 (not /r/), no
//      borrow rule #2 (no receiver) → m.Realm stays victim
//   7. EvilWrite body: i.Field = ... writes /r/-victim data with
//      m.Realm = victim → readonly check sees crlm == storage → no
//      panic → WRITE SUCCEEDS
//
// This is the SAME known-open Apply pattern, just with defer. The
// defer path doesn't add any protection.
//
// Expected (matching the known-open pattern documented in
// other launder tests): the embedded /p/.Immutable is mutated.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadPtr())

	inner := launderrvictim.GetWithPtr().Inner
	inner.DeferApply(launderattack.EvilWrite)

	println("after: ", launderrvictim.ReadPtr())
}

// Output:
// before: ptr-orig

// Error:
// cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): i<VPBlock(1,0)>.Field
