// Package v3acanary is a v3a-style fixture realm — NOT a reusable
// Ownable library. It exercises the v3a ACL shape end-to-end so the
// filetest can assert the security claim. Function names are
// deliberately verb-prefixed (Seed/Reassign/Clear) instead of
// Init/TransferOwnership/DropOwnership so this file isn't mistaken
// for the canonical owner pattern. The canonical owner pattern, if
// ever needed, belongs under /p/.
//
// Compare to v2's Ownable, sketched in spec form:
//
//	func (o *Ownable) TransferOwnership(_ int, rlm realm, newOwner address) error {
//	    if !rlm.IsCurrent() { return ErrUnauthorized }
//	    caller := rlm.Previous().Address()
//	    if !o.OwnedBy(caller) { return ErrUnauthorized }
//	    o.owner = newOwner
//	    return nil
//	}
//
// v3a equivalent below: no `_ int` slot, no `rlm realm` parameter,
// no IsCurrent check. Identity comes from the VM via runtime.Caller().
// The omarsy forgery class has no surface because there's no
// realm-typed value passing across the boundary.
package v3acanary

import "chain/runtime"

var (
	owner  address
	seeded bool
)

// Seed records the calling realm as the principal. Idempotent. The
// v2-style name would be Init; we avoid that here so the fixture
// can't be mistaken for an Ownable.
func Seed() {
	if seeded {
		return
	}
	owner = runtime.Caller().Address()
	seeded = true
}

// Owner returns the current principal.
func Owner() address {
	return owner
}

// Reassign moves the principal to newOwner. Authorization: the caller
// (per runtime.Caller()) must equal the current owner.
//
// No `cur realm` / `rlm realm` parameter — the v3a pattern.
func Reassign(newOwner address) {
	caller := runtime.Caller().Address()
	if caller != owner {
		panic("not owner")
	}
	if !newOwner.IsValid() {
		panic("invalid address")
	}
	owner = newOwner
}

// Clear unsets the principal. Same authorization shape.
func Clear() {
	caller := runtime.Caller().Address()
	if caller != owner {
		panic("not owner")
	}
	owner = address("")
}
