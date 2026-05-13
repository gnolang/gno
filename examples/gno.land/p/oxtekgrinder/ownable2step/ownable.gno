package ownable2step

import "chain"

const OwnershipTransferEvent = "OwnershipTransfer"

// Ownable2Step is a two-step ownership transfer package.
// It allows the current owner to set a new owner; the new owner must
// accept ownership before the transfer completes.
//
// All authority-checking methods take an explicit `caller address`
// parameter. Callers extract the appropriate address from their `cur`
// realm at the call site:
//
//	o.TransferOwnership(cur.Address(), newOwner)             // realm-as-owner
//	o.TransferOwnership(cur.Previous().Address(), newOwner)  // caller-of-realm-as-owner
//
// XXX Implement using Ownable instead of replicating it.
type Ownable2Step struct {
	owner        address
	pendingOwner address
}

// NewWithAddress constructs an Ownable2Step with the given owner. This
// is the only constructor — the previous NewWithOrigin / NewWithCurrent
// sugar baked in runtime walks; the realm using this package now picks
// the owner address explicitly (e.g. cur.Previous().Address() after
// verifying cur.Previous().IsUserCall() in init).
func NewWithAddress(addr address) *Ownable2Step {
	return &Ownable2Step{
		owner:        addr,
		pendingOwner: "",
	}
}

// TransferOwnership initiates the transfer of ownership to newOwner by
// setting the PendingOwner. caller must equal the current owner.
func (o *Ownable2Step) TransferOwnership(caller, newOwner address) error {
	if caller != o.owner {
		return ErrUnauthorized
	}
	if !newOwner.IsValid() {
		return ErrInvalidAddress
	}
	o.pendingOwner = newOwner
	return nil
}

// AcceptOwnership accepts the pending ownership transfer. caller must
// equal the pending owner.
func (o *Ownable2Step) AcceptOwnership(caller address) error {
	if o.pendingOwner.String() == "" {
		return ErrNoPendingOwner
	}
	if caller != o.pendingOwner {
		return ErrPendingUnauthorized
	}
	o.owner = o.pendingOwner
	o.pendingOwner = ""
	return nil
}

// DropOwnership removes the owner, disabling any owner-related actions.
// caller must equal the current owner.
func (o *Ownable2Step) DropOwnership(caller address) error {
	if caller != o.owner {
		return ErrUnauthorized
	}
	prevOwner := o.owner
	o.owner = ""
	o.pendingOwner = ""
	chain.Emit(
		OwnershipTransferEvent,
		"from", prevOwner.String(),
		"to", "",
	)
	return nil
}

// Owner returns the current owner address.
func (o *Ownable2Step) Owner() address {
	return o.owner
}

// PendingOwner returns the pending owner address (empty if none).
func (o *Ownable2Step) PendingOwner() address {
	return o.pendingOwner
}

// OwnedBy reports whether the given address is the current owner.
func (o *Ownable2Step) OwnedBy(addr address) bool {
	return addr == o.owner
}

// AssertOwnedBy panics with ErrUnauthorized if addr is not the owner.
func (o *Ownable2Step) AssertOwnedBy(addr address) {
	if addr != o.owner {
		panic(ErrUnauthorized)
	}
}
