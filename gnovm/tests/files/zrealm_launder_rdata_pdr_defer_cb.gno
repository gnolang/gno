// PKGPATH: gno.land/r/launderrh_pdr_defer_cb
package launderrh_pdr_defer_cb

// Vector 2: /r/-victim defers an attacker-supplied callback. At
// frame-pop time, the defer's PushFrameCall fires the borrow rules
// on the deferred function (which is /r/-attacker-declared).
//
// Trace:
//  1. main → m.Realm = /r/launderrh_pdr_defer_cb
//  2. DeferCallback(h) → borrow rule #1 → m.Realm = /r/launderrvictim
//  3. body: `defer h(gImm)` schedules h on defer stack
//  4. body returns → doOpReturnCallDefers picks up deferred h
//  5. PushFrameCall for h: h is /r/-attacker-declared → borrow rule #1
//     → m.Realm = /r/launderrh_pdr_defer_cb
//  6. h body: `i.Field = ...` → readonly check fires (target
//     object stamped /r/launderrvictim, m.Realm = attacker)
//
// Expected: the readonly panic fires at the write, same as the
// non-defer Apply case. The defer path doesn't reset the borrow
// rules.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadImm())

	launderrvictim.DeferCallback(func(i *launderrvictim.Immutable) {
		i.Field = "pwnd-via-defer-cb"
	})

	println("after: ", launderrvictim.ReadImm())
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): i<VPBlock(1,0)>.Field
