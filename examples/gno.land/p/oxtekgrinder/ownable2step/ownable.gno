package ownable2step

import "chain"

const OwnershipTransferEvent = "OwnershipTransfer"

// Ownable2Step is a two-step ownership transfer package
// It allows the current owner to set a new owner and the new owner will need to accept the ownership before it is transferred
// XXX Implement using Ownable instead of replicating it.
type Ownable2Step struct {
	owner        address
	pendingOwner address
}

func New(_ int, rlm realm) *Ownable2Step {
	return &Ownable2Step{
		owner:        rlm.Address(),
		pendingOwner: "",
	}
}

func NewWithOrigin(_ int, rlm realm) *Ownable2Step {
	prev := rlm.Previous()
	if !prev.IsUserCall() {
		panic("NewWithOrigin() should be called from init() where the previous realm is the EOA origin")
	}
	return &Ownable2Step{
		owner: prev.Address(),
	}
}

func NewWithAddress(addr address) *Ownable2Step {
	return &Ownable2Step{
		owner:        addr,
		pendingOwner: "",
	}
}

// TransferOwnership initiate the transfer of the ownership to a new address by setting the PendingOwner
func (o *Ownable2Step) TransferOwnership(_ int, rlm realm, newOwner address) error {
	if !o.OwnedByCurrent(0, rlm) {
		return ErrUnauthorized
	}
	if !newOwner.IsValid() {
		return ErrInvalidAddress
	}

	o.pendingOwner = newOwner
	return nil
}

// AcceptOwnership accepts the pending ownership transfer
func (o *Ownable2Step) AcceptOwnership(_ int, rlm realm) error {
	if o.pendingOwner.String() == "" {
		return ErrNoPendingOwner
	}
	if rlm.Address() != o.pendingOwner {
		return ErrPendingUnauthorized
	}

	o.owner = o.pendingOwner
	o.pendingOwner = ""

	return nil
}

// DropOwnership removes the owner, effectively disabling any owner-related actions
// Top-level usage: disables all only-owner actions/functions,
// Embedded usage: behaves like a burn functionality, removing the owner from the struct
func (o *Ownable2Step) DropOwnership(_ int, rlm realm) error {
	if !o.OwnedByCurrent(0, rlm) {
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

// Owner returns the owner address from Ownable
func (o *Ownable2Step) Owner() address {
	return o.owner
}

// PendingOwner returns the pending owner address from Ownable2Step
func (o *Ownable2Step) PendingOwner() address {
	return o.pendingOwner
}

// OwnedByCurrent checks if the caller of the function is the Realm's owner
func (o *Ownable2Step) OwnedByCurrent(_ int, rlm realm) bool {
	return rlm.Address() == o.owner
}

// AssertOwnedByCurrent panics if the caller is not the owner
func (o *Ownable2Step) AssertOwnedByCurrent(_ int, rlm realm) {
	if rlm.Address() != o.owner {
		panic(ErrUnauthorized)
	}
}

// OwnedByPrevious checks if the caller of the function is the Realm's owner
func (o *Ownable2Step) OwnedByPrevious(_ int, rlm realm) bool {
	return rlm.Previous().Address() == o.owner
}

// AssertOwnedByPrevious panics if the caller is not the owner
func (o *Ownable2Step) AssertOwnedByPrevious(_ int, rlm realm) {
	if rlm.Previous().Address() != o.owner {
		panic(ErrUnauthorized)
	}
}
