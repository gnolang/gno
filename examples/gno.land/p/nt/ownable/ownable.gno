package ownable

import "chain"

const OwnershipTransferEvent = "OwnershipTransfer"

// Ownable is meant to be used as a top-level object to make your contract
// ownable OR being embedded in a Gno object to manage per-object ownership.
// Ownable is safe to export as a top-level object.
//
// Authority-mutating methods (TransferOwnership, DropOwnership) take
// (_ int, rlm realm). The caller threads its own cur; the method
// asserts rlm.IsCurrent() and identifies the principal as
// rlm.Previous().Address() — which must equal the current owner.
//
//	o.TransferOwnership(0, cur, newOwner)
//	o.DropOwnership(0, cur)
//
// Read methods (OwnedBy, AssertOwnedBy) keep the bare-address shape;
// callers extract the address themselves (e.g. cur.Previous().Address()).
type Ownable struct {
	owner address
}

// NewWithAddress creates an Ownable with the given address as owner.
// This is the only constructor — the previous New/NewWithOrigin/
// NewWithAddressByPrevious sugar baked runtime walks and an auth-mode
// flag into the struct; the realm using this package now picks the
// owner address explicitly (e.g. cur.Previous().Address() after
// verifying cur.Previous().IsUserCall() in init).
func NewWithAddress(addr address) *Ownable {
	return &Ownable{
		owner: addr,
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

// TransferOwnership transfers ownership of the Ownable to newOwner. rlm
// must be the caller's own captured cur (asserted via rlm.IsCurrent()).
// The principal is rlm.Previous().Address() — the realm that crossed
// into the caller — which must equal the current owner.
//
// IsCurrent + rlm.Previous() makes the principal unforgeable: an
// attacker calling TransferOwnership on a foreign Ownable cannot supply
// an arbitrary caller address; rlm comes from a runtime-validated
// crossing frame.
func (o *Ownable) TransferOwnership(_ int, rlm realm, newOwner address) error {
	if !rlm.IsCurrent() {
		return ErrUnauthorized
	}
	caller := rlm.Previous().Address()
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

// DropOwnership removes the owner, disabling any owner-related actions.
// rlm must be the caller's own captured cur; rlm.Previous().Address()
// must equal the current owner.
func (o *Ownable) DropOwnership(_ int, rlm realm) error {
	if !rlm.IsCurrent() {
		return ErrUnauthorized
	}
	caller := rlm.Previous().Address()
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

// Owner returns the owner address.
func (o *Ownable) Owner() address {
	if o == nil {
		return address("")
	}
	return o.owner
}
