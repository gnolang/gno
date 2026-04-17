// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"chain/runtime"

	"gno.land/r/tests/vm/crossrealm_e"
)

func main() {
	println("owner:", crossrealm_e.GetOwner())
	attacker := runtime.CurrentRealm().Address()
	println("attacker: ", attacker)

	func() {
		defer func() {
			r := recover()
			if r != nil {
				println("exception caught: ", r.(string))
			}
		}()
		crossrealm_e.SetOwner(runtime.CurrentRealm().Address())
	}()

	owner := crossrealm_e.GetOwner()
	println("owner == attacker: ", owner == attacker)

	crossrealm_e.TransferToken(cross)
}

// before fix:
// owner: g1dao_address_here
// attacker:  g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan
// exception caught:  cannot modify external-realm or non-realm object
// owner == attacker:  true
// ===send token to:  g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan

// after fix:

// Output:
// owner: g1dao_address_here
// attacker:  g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan
// exception caught:  cannot directly modify readonly tainted object (use a method or crossing function): owner<~VPBlock(3,1)>
// owner == attacker:  false

// Error:
// unauthorized
