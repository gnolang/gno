// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Demonstrates that assign+recover corrupts state even without
// OriginCaller or fund theft. An attacker flips a governance
// vote from "rejected" to "approved" via pure state corruption.
//
// NOTE: this corruption is within-tx only — the store layer
// does not persist it because DidUpdate panics before MarkDirty.

import (
	"gno.land/r/tests/vm/crossrealm_d"
)

func main() {
	// 1. Verify initial state: proposal is rejected (10 yes, 90 no).
	result := crossrealm_d.GetResult(cross)
	println("before:", result)

	// 2. Corrupt the tally via assign+recover.
	//    SetVotes is non-crossing → cross-realm guard panics.
	//    But the mutation lands before the panic.
	func() {
		defer func() {
			r := recover()
			if r != nil {
				println("panic caught:", r.(string))
			}
		}()
		crossrealm_d.SetVotes(99, 1)
	}()

	// 3. Check result after corruption.
	result = crossrealm_d.GetResult(cross)
	println("after:", result)
}

// Output:
// before: rejected
// panic caught: cannot modify external-realm or non-realm object
// after: approved
