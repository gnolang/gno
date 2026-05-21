// PKGPATH: gno.land/r/v3acanaryhost
package v3acanaryhost

import (
	"chain/runtime"

	"gno.land/r/tests/vm/v3acanary"
)

// Phase B-min end-to-end: exercise the v3a-style canary realm from
// /r/v3acanaryhost. The canary uses runtime.Caller() for ACL — no
// `cur realm` / `rlm realm` parameters, no `cross` keyword at the
// call site. Phase B-min's auto-cross walk makes the call
// /r/v3acanaryhost → /r/v3acanary register as a realm transition,
// so Caller() inside Init() resolves to /r/v3acanaryhost (the
// immediate caller), not the EOA.

func main() {
	// Bootstrap: seed the principal via Caller() inside Seed.
	v3acanary.Seed()

	// Security claim: the principal is the calling realm
	// (/r/v3acanaryhost), NOT the EOA. Without auto-cross on
	// /r/-exported calls, Caller() would have returned the EOA and
	// any other realm could launder calls through itself to act as
	// that EOA — the exact pattern v3a closes.
	host := runtime.Self().Address()
	eoa := runtime.Origin().Address()
	println("owner == host:", v3acanary.Owner() == host)
	println("owner != eoa:", v3acanary.Owner() != eoa)

	// Owner-authenticated reassignment: succeeds because Caller()
	// inside Reassign also resolves to /r/v3acanaryhost = owner.
	newOwner := address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
	v3acanary.Reassign(newOwner)
	println("owner after reassign:", v3acanary.Owner() == newOwner)

	// After reassign, /r/v3acanaryhost is no longer owner; the
	// canary's ACL should reject. Cross-realm panics become aborts
	// under v2 Phase 3, so we use revive() to catch.
	r := revive(func() { v3acanary.Reassign(host) })
	println("post-reassign reject:", r != nil)
}

// Output:
// owner == host: true
// owner != eoa: true
// owner after reassign: true
// post-reassign reject: true
