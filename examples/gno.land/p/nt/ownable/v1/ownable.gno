// Package ownable (v1) is the v3a-aligned successor to /p/nt/ownable/v0.
//
// Differences from v0:
//   - No `_ int, rlm realm` parameter shape on mutating methods.
//   - No `rlm.IsCurrent()` check (no rlm value to forge).
//   - Caller identity sourced from runtime.Caller() instead of
//     rlm.Previous().
//
// The omarsy forgery class has no surface here: there is no
// realm-typed parameter; the principal is queried by the VM from the
// live stack and returned as a concrete runtime.Realm struct that
// user code cannot construct in disguise.
//
// Usage:
//
//	o := ownable.NewWithAddress(alice)
//	o.TransferOwnership(bob)   // caller must be alice; checked via runtime.Caller()
//	o.DropOwnership()           // caller must be alice
//
// See /p/nt/ownable/v0 for the v2-era equivalent that uses cur/rlm
// parameters. v1 is preferred for new code; v0 is preserved during
// the migration window.
package ownable

import (
	"chain"
	"chain/runtime"
)

const OwnershipTransferEvent = "OwnershipTransfer"

// Ownable holds an owner address. Safe to export as a top-level
// object or embed in a Gno object for per-object ownership.
type Ownable struct {
	owner address
}

// NewWithAddress creates an Ownable with the given address as owner.
func NewWithAddress(addr address) *Ownable {
	return &Ownable{
		owner: addr,
	}
}

// NewWithCaller creates an Ownable with the calling realm's caller as
// owner. Equivalent to NewWithAddress(runtime.Caller().Address()) but
// expresses intent more clearly at construction sites where the
// canonical "owner is whoever set this up" pattern applies.
func NewWithCaller() *Ownable {
	return &Ownable{
		owner: runtime.Caller().Address(),
	}
}

// OwnedBy reports whether addr is the current owner.
func (o *Ownable) OwnedBy(addr address) bool {
	if o == nil {
		return false
	}
	return addr == o.owner
}

// AssertOwnedBy panics with ErrUnauthorized if addr is not the owner.
func (o *Ownable) AssertOwnedBy(addr address) {
	if !o.OwnedBy(addr) {
		panic(ErrUnauthorized)
	}
}

// TransferOwnership transfers ownership to newOwner. Authorization:
// runtime.Caller() must equal the current owner.
func (o *Ownable) TransferOwnership(newOwner address) error {
	caller := runtime.Caller().Address()
	if !o.OwnedBy(caller) {
		return ErrUnauthorized
	}
	if !newOwner.IsValid() {
		return ErrInvalidAddress
	}
	prevOwner := o.owner
	o.owner = newOwner
	chain.Emit(
		OwnershipTransferEvent,
		"from", prevOwner.String(),
		"to", newOwner.String(),
	)
	return nil
}

// DropOwnership clears the owner, disabling any owner-gated actions.
// Authorization: runtime.Caller() must equal the current owner.
func (o *Ownable) DropOwnership() error {
	caller := runtime.Caller().Address()
	if !o.OwnedBy(caller) {
		return ErrUnauthorized
	}
	prevOwner := o.owner
	o.owner = ""
	chain.Emit(
		OwnershipTransferEvent,
		"from", prevOwner.String(),
		"to", "",
	)
	return nil
}

// Owner returns the current owner address.
func (o *Ownable) Owner() address {
	if o == nil {
		return address("")
	}
	return o.owner
}
