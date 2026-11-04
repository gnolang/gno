// PKGPATH: gno.land/r/example
//
// Executes the contract-identity gate from a real realm.
//
// Two reasons this cannot be a plain unit test in p/moul/authz:
//
//  1. Example_* functions that take parameters and have no "// Output:"
//     are never run by `gno test` — a sentinel panic inside one still
//     reports ok. So the package's headline example, the shape readers
//     copy, is invisible to CI unless something else runs it.
//  2. The gate cannot be simulated from inside the package. The VM mints
//     a crossing frame's realm from the CALLEE's declaring package, so a
//     closure declared in p/moul/authz always presents that package's
//     address no matter what testing.SetRealm says. Only a realm can
//     present its own.
//
// (2) is the same mechanism as the capability leak, seen from the other
// side: it is what makes `rlm.Address()` inside a crossing frame
// unconditionally the realm's own address — and therefore what makes
// `NewContractAuthority(ownPath)` + `DoByCurrent` a comparison of the
// realm to itself that can never reject.
//
// WHAT THIS FILE PINS. Both halves go red if the identity gate is
// reverted to its pre-fix `proposer: NewAutoAcceptAuthority()` default:
//
//   - a principal that is NOT the contract is refused;
//   - the contract itself is accepted, so the gate does not lock out the
//     legitimate driver.
//
// It deliberately does NOT demonstrate `DoByCurrent` on an own-path
// authority. That pairing is the anti-pattern NewContractAuthority's
// godoc forbids: it succeeds here and it would equally succeed for a
// hostile caller, so asserting it green would document the hole as the
// recommended shape. A filetest is confined to one realm and so cannot
// exhibit the foreign caller itself; the package-level negatives
// (TestForeignFrameCannotDriveContractAuthority,
// TestContractPathIsolation) and r/gnops/valopers' filetests cover that.
package example

import (
	"chain"

	"gno.land/p/moul/authz/v0"
)

var auth = authz.NewWithAuthority(
	authz.NewContractAuthority(
		"gno.land/r/example", // this realm's own path
		func(_ string, action authz.PrivilegedAction) error { return action() },
	),
)

var updated string

// UpdateParams is the exported crossing entrypoint of the realm.
//
// DoByPrevious, not DoByCurrent: the principal is whoever crossed in, so
// an unrelated realm calling this presents its own address and is
// refused. Under DoByCurrent the principal would be this realm's own
// address for every caller, and this entrypoint would be an
// unauthenticated privileged write.
//
// Note what must NOT be added next to it: a function returning this
// closure, or any other value carrying this frame. That is the leak.
func UpdateParams(cur realm, v string) error {
	return auth.DoByPrevious(0, cur, "update_params", func() error {
		updated = v
		return nil
	})
}

func main(cur realm) {
	// A principal that is not the contract is refused, and the action
	// does not run. This is the assertion the pre-fix AutoAccept default
	// could not make.
	outsider := chain.PackageAddress("gno.land/r/outsider")
	err := auth.Authority().Authorize(outsider, "update_params", func() error {
		updated = "SHOULD-NOT-LAND"
		return nil
	})
	println("outsider err:", err)
	println("updated:", updated)

	// The contract itself is accepted: main and UpdateParams are both in
	// gno.land/r/example, so cur.Previous() is this realm.
	println("contract err:", UpdateParams(cross(cur), "governed-value"))
	println("updated:", updated)
}

// Output:
// outsider err: unauthorized
// updated:
// contract err: undefined
// updated: governed-value
