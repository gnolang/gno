// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Verifies that cross-realm assign+recover does NOT corrupt state.
//
// crossrealm_d has a package-level counter (initially 100) with both
// a non-crossing setter (SetCounter) and a crossing setter
// (SetCounterCrossing). This test calls the non-crossing setter from
// another realm, recovers the panic, then checks that the state is
// unchanged.
//
// Impact without the fix: the non-crossing SetCounter(0) would land
// in memory before DidUpdate panics. After recover(), the counter is
// 0 in memory but 100 in persistence. A subsequent crossing call like
// DoubleCounter would read 0 and produce 0 instead of 200 — the
// realm acts on state that was never committed.
//
// The fix: PopAsPointer checks readonly BEFORE the assignment, so
// the counter never changes in memory.

import (
	"gno.land/r/tests/vm/crossrealm_d"
)

func main() {
	// 1. Initial state.
	println("counter:", crossrealm_d.GetCounter(cross)) // 100

	// 2. Non-crossing set → panics, recover catches it.
	func() {
		defer func() {
			r := recover()
			if r != nil {
				println("panic caught:", r.(string))
			}
		}()
		crossrealm_d.SetCounter(0)
	}()

	// 3. Counter must still be 100.
	println("counter after set+recover:", crossrealm_d.GetCounter(cross))

	// 4. A crossing call that depends on counter.
	//    Without the fix: counter=0 in memory → DoubleCounter returns 0.
	//    With the fix: counter=100 → DoubleCounter returns 200.
	result := crossrealm_d.DoubleCounter(cross)
	println("double:", result)

	// 5. Verify the crossing setter works correctly for comparison.
	crossrealm_d.SetCounterCrossing(cross, 50)
	println("after crossing set:", crossrealm_d.GetCounter(cross)) // 50
}

// Output:
// counter: 100
// panic caught: cannot directly modify readonly tainted object (use a method or crossing function): counter<~VPBlock(3,0)>
// counter after set+recover: 100
// double: 200
// after crossing set: 50
