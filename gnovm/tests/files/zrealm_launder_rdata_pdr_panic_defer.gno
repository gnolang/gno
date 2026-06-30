// PKGPATH: gno.land/r/launderrh_pdr_panic_defer
package launderrh_pdr_panic_defer

// Vector 3: /r/-victim defers an attacker callback, then panics.
// The deferred attacker callback runs during panic unwinding.
//
// PanicAfterPushDefer:
//   defer h(gImm)
//   panic("victim-induced panic")
//
// Question: does the defer's PushFrameCall still apply borrow rule #1 to
// borrow m.Realm to attacker, even though it's running during
// panic unwind? The doOpReturnCallDefers path uses the same
// m.PushFrameCall(..., true) — so the borrow logic should be
// identical.
//
// Expected: the write inside h triggers a readonly check (because
// h borrow rule #1 borrows m.Realm to attacker). The readonly panic
// REPLACES the original "victim-induced panic" exception in m.Exception.
// Then unwinding crosses the realm boundary → unhandled panic abort.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadImm())

	launderrvictim.PanicAfterPushDefer(func(i *launderrvictim.Immutable) {
		i.Field = "pwnd-via-panic-defer"
	})

	println("after: ", launderrvictim.ReadImm())
}

// Error:
// victim-induced panic
// 	cannot directly modify readonly tainted object (use a method or crossing function): i<VPBlock(1,0)>.Field
