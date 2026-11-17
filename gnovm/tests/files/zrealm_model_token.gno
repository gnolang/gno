// PKGPATH: gno.land/r/thief
// The token scenario of docs/resources/gno-interrealm-model.md: the crossing
// gate moves the caller's own tokens; reading the ledger is fine; a library
// method on the token realm's ledger, called by another realm, is refused.
package thief

import "gno.land/r/tests/vm/modeltoken"

const (
	alice = address("g1alice")
	bob   = address("g1bob")
)

func main(cur realm) {
	modeltoken.Transfer(cross(cur), bob, 10)
	println("bob:", modeltoken.Ledger().Balance(bob))
	modeltoken.Ledger().Move(alice, "g1thief", 1000000000)
}

// Output:
// bob: 10

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): l<VPBlock(1,0)>.bal[from<VPBlock(1,1)>]
