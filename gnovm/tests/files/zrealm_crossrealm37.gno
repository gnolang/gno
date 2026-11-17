// PKGPATH: gno.land/r/crossrealm
package crossrealm

// SetOwner is non-crossing, so called from another realm crossrealm_e runs
// read-only and the hijack is refused; the owner check then holds.

import (
	"chain/runtime/unsafe"

	"gno.land/r/tests/vm/crossrealm_e"
)

func main(cur realm) {
	println("owner:", crossrealm_e.GetOwner())
	attacker := unsafe.CurrentRealm().Address()
	println("attacker: ", attacker)

	// recover() cannot catch a panic crossing a realm boundary; revive can.
	println("exception caught:", try(func() { crossrealm_e.SetOwner(unsafe.CurrentRealm().Address()) }))

	owner := crossrealm_e.GetOwner()
	println("owner == attacker: ", owner == attacker)

	println("transfer:", revive(func() { crossrealm_e.TransferToken(cross(cur)) }))
}

// try runs fn and returns what it panicked with: a refused write in a
// view-only frame is recovered by the caller, not revived.
func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// owner: g1dao_address_here
// attacker:  g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan
// exception caught: gno.land/r/tests/vm/crossrealm_e's code was entered without crossing and is read-only (call its crossing function): owner<~VPBlock(3,1)>
// owner == attacker:  false
// transfer: unauthorized
