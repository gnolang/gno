// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Demonstrates the accidental footgun scenario:
//
// - Callee (crossrealm_d) exports SetCounter without `cur realm` —
//   an oversight, perhaps a helper that was meant to be unexported.
//
// - Caller does a batch update (DoubleCounter + SetCounter) wrapped
//   in recover because DoubleCounter might overflow. The readonly
//   panic from SetCounter slips through the same recover accidentally.
//
// Without the pre-mutation fix: DoubleCounter succeeds (100→200),
// then SetCounter(0) lands in memory, the readonly panic is caught
// by the recover meant for overflow, and the caller proceeds with
// counter=0 instead of 200.
//
// With the fix: DoubleCounter succeeds (100→200), SetCounter(0)
// is blocked before mutation, counter stays 200.

import (
	"strings"

	"gno.land/r/tests/vm/crossrealm_d"
)

func main() {
	println("counter:", crossrealm_d.GetCounter(cross))

	// Caller does a batch update: double the counter, then reset it.
	// DoubleCounter can panic (e.g., overflow), so the caller wraps
	// the whole batch in a recover to handle that gracefully.
	func() {
		defer func() {
			r := recover()
			if r != nil {
				msg := r.(string)
				if strings.Contains(msg, "overflow") {
					println("handled: overflow, skipping batch")
				}
				// Any other panic (including readonly) is silently
				// swallowed — the caller didn't anticipate it.
			}
		}()
		crossrealm_d.DoubleCounter(cross) // might overflow — the reason for recover
		crossrealm_d.SetCounter(0)        // oversight: non-crossing, panics with readonly
	}()

	// Caller assumes the batch succeeded (no overflow).
	// They expect counter=0, but DoubleCounter succeeded (100→200)
	// and SetCounter(0) was silently blocked. Counter is now 200.
	val := crossrealm_d.GetCounter(cross)
	println("counter after:", val)
	println("calculation:", val*3)
}

// XXX, consequence of the incorrect state, if no pre-mutation guard, the calculation will be 0, thus e.g. :
//   - Emit an event: chain.Emit("result", "0") — persisted, indexed, consumed by off-chain systems
//   - Pass it to another crossing call: otherRealm.Record(cross, 0) — persisted in another realm's state
//   - Use it for a banker transfer: SendCoins(... amount: 0) — no funds move when they should

//   All of these are committed. But the callee's counter reverts to 200 after the tx. Now on-chain state is inconsistent
//   across realms — the event says 0, another realm recorded 0, but the source of truth is 200. And there's no trace of what
//   went wrong.

// Output:
// counter: 100
// counter after: 200
// calculation: 600
