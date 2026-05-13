package ownable

import "chain"

const OwnershipTransferEvent = "OwnershipTransfer"

// Ownable is meant to be used as a top-level object to make your contract
// ownable OR being embedded in a Gno object to manage per-object ownership.
// Ownable is safe to export as a top-level object.
//
// All authority-checking methods take an explicit `caller address`
// parameter. Callers extract the appropriate address from their `cur`
// realm at the call site:
//
//	o.TransferOwnership(cur.Address(), newOwner)             // realm-as-owner
//	o.TransferOwnership(cur.Previous().Address(), newOwner)  // caller-of-realm-as-owner
//
// This replaces the previous current-vs-previous auth mode flag baked
// into the struct — the choice now lives at each call site, which makes
// the authority surface explicit at every use rather than at construction.
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

// TransferOwnership transfers ownership of the Ownable to newOwner.
// caller must equal the current owner.
func (o *Ownable) TransferOwnership(caller, newOwner address) error {
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
// caller must equal the current owner.
func (o *Ownable) DropOwnership(caller address) error {
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
