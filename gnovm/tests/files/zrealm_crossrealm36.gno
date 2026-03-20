// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Verifies that cross-realm admin assignment via assign+recover
// does NOT corrupt state. The readonly check in PopAsPointer fires
// BEFORE the mutation, so admin remains unchanged after recover().

import (
	"chain/runtime"

	"gno.land/r/tests/vm/crossrealm_d"
)

func main() {
	adminBefore := crossrealm_d.GetAdmin(cross)
	println("admin before:", adminBefore)

	// Attempt to corrupt admin to this realm's address via assign+recover.
	myAddr := runtime.CurrentRealm().Address()
	func() {
		defer func() {
			r := recover()
			if r != nil {
				println("panic caught:", r.(string))
			}
		}()
		crossrealm_d.SetAdmin(myAddr)
	}()

	// Verify admin is unchanged (fix prevents corruption).
	adminAfter := crossrealm_d.GetAdmin(cross)
	println("admin after:", adminAfter)
	println("admin matches attacker:", adminAfter == myAddr)

	println("realm state intact")
}

// Output:
// admin before:
// panic caught: cannot directly modify readonly tainted object (w/o method): admin<~VPBlock(3,6)>
// admin after:
// admin matches attacker: false
// realm state intact
