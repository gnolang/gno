// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Verifies that cross-realm assign+recover does NOT corrupt state.
// The readonly check in PopAsPointer fires BEFORE the mutation,
// so recover() catches the panic but no state change occurs.

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
// panic caught: cannot directly modify readonly tainted object (w/o method): yesVotes<~VPBlock(3,0)>
// after: rejected
