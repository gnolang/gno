// Package v3acanary is a v3a-style reference realm. It demonstrates
// the ACL pattern using runtime.Caller() instead of v2's `cur realm` /
// `rlm realm` parameters.
//
// Compare to v2's Ownable:
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
	owner       address
	initialized bool
)

// Init seeds the owner to the calling realm's caller. Idempotent.
// Demonstrates: the EOA that issued the MsgCall is the principal we
// record, sourced via runtime.Caller() inside a non-crossing helper.
func Init() {
	if initialized {
		return
	}
	owner = runtime.Caller().Address()
	initialized = true
}

// Owner returns the current owner.
func Owner() address {
	return owner
}

// TransferOwnership transfers to newOwner. Authorization: the caller
// (per runtime.Caller()) must equal the current owner.
//
// No `cur realm` / `rlm realm` parameter — the v3a pattern.
func TransferOwnership(newOwner address) {
	caller := runtime.Caller().Address()
	if caller != owner {
		panic("not owner")
	}
	if !newOwner.IsValid() {
		panic("invalid address")
	}
	owner = newOwner
}

// DropOwnership clears the owner. Same authorization shape.
func DropOwnership() {
	caller := runtime.Caller().Address()
	if caller != owner {
		panic("not owner")
	}
	owner = address("")
}
