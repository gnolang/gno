// PKGPATH: gno.land/r/v3acanaryhost
package v3acanaryhost

import (
	"chain/runtime"

	"gno.land/r/tests/vm/v3acanary"
)

// Phase B starter: end-to-end exercise of the v3a-style canary realm.
// /r/v3acanary uses runtime.Caller() for ACL — no `cur realm` /
// `rlm realm` parameters. We drive it from /r/v3acanaryhost via
// regular method calls; the canary's ACL checks should resolve to
// /r/v3acanaryhost's caller (the EOA) as the principal.

func main(cur realm) {
	// Bootstrap: seed the owner.
	v3acanary.Init()

	// The seeded owner should be the EOA (our MsgCall signer).
	eoa := runtime.PreviousRealm().Address()
	println("owner == eoa:", v3acanary.Owner() == eoa)

	// Owner-authenticated transfer succeeds (caller is the EOA = owner).
	newOwner := address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
	v3acanary.TransferOwnership(newOwner)
	println("owner after transfer:", v3acanary.Owner() == newOwner)

	// Subsequent call: the original EOA is no longer owner, so the
	// ACL should reject. Cross-realm panics become aborts under v2
	// Phase 3, so we use revive() to catch.
	r := revive(func() { v3acanary.TransferOwnership(eoa) })
	println("post-transfer reject:", r != nil)
}

// Output:
// owner == eoa: true
// owner after transfer: true
// post-transfer reject: true
