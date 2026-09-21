// PKGPATH: gno.land/r/launderrh_pdr_defer_apply
package launderrh_pdr_defer_apply

// Vector 4: /r/-victim DeferApplyHook defers a call to ApplyHook
// from within victim's frame. The deferred ApplyHook in turn
// invokes the attacker callback.
//
// Trace:
//  1. main → m.Realm = /r/launderrh_pdr_defer_apply
//  2. DeferApplyHook(h) → borrow rule #1 → m.Realm = /r/launderrvictim
//  3. body: `defer ApplyHook(h)` schedules ApplyHook on defer stack
//  4. body returns → doOpReturnCallDefers picks up deferred ApplyHook
//  5. PushFrameCall for ApplyHook: /r/-declared → borrow rule #1
//     → m.Realm STAYS at /r/launderrvictim (already)
//  6. ApplyHook body: h(gImm) → another PushFrameCall for h
//     → borrow rule #1 → m.Realm = /r/launderrh_pdr_defer_apply (attacker)
//  7. h body: i.Field = ... → readonly check fires
//
// Expected: same outcome as direct ApplyHook — the write is
// blocked. The deferred routing of ApplyHook doesn't bypass any
// borrow rule.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadImm())

	launderrvictim.DeferApplyHook(func(i *launderrvictim.Immutable) {
		i.Field = "pwnd-via-defer-apply"
	})

	println("after: ", launderrvictim.ReadImm())
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): i<VPBlock(1,0)>.Field
