package ownable2step

import (
	"chain"
	"chain/runtime"
)

const OwnershipTransferEvent = "OwnershipTransfer"

// Ownable2Step is a two-step ownership transfer package
// It allows the current owner to set a new owner and the new owner will need to accept the ownership before it is transferred
// XXX Implement using Ownable instead of replicating it.
type Ownable2Step struct {
	owner        address
	pendingOwner address
}

func New() *Ownable2Step {
	return &Ownable2Step{
		owner:        runtime.CurrentRealm().Address(),
		pendingOwner: "",
	}
}

func NewWithOrigin() *Ownable2Step {
	origin := runtime.OriginCaller()
	previous := runtime.PreviousRealm()
	if origin != previous.Address() {
		panic("NewWithOrigin() should be called from init() where std.PreviousRealm() is origin")
	}
	return &Ownable2Step{
		owner: origin,
	}
}

func NewWithAddress(addr address) *Ownable2Step {
	return &Ownable2Step{
		owner:        addr,
		pendingOwner: "",
	}
}

// TransferOwnership initiate the transfer of the ownership to a new address by setting the PendingOwner
func (o *Ownable2Step) TransferOwnership(newOwner address) error {
	if !o.OwnedByCurrent() {
		return ErrUnauthorized
	}
	if !newOwner.IsValid() {
		return ErrInvalidAddress
	}

	o.pendingOwner = newOwner
	return nil
}

// AcceptOwnership accepts the pending ownership transfer
func (o *Ownable2Step) AcceptOwnership() error {
	if o.pendingOwner.String() == "" {
		return ErrNoPendingOwner
	}
	if runtime.CurrentRealm().Address() != o.pendingOwner {
		return ErrPendingUnauthorized
	}

	o.owner = o.pendingOwner
	o.pendingOwner = ""

	return nil
}

// DropOwnership removes the owner, effectively disabling any owner-related actions
// Top-level usage: disables all only-owner actions/functions,
// Embedded usage: behaves like a burn functionality, removing the owner from the struct
func (o *Ownable2Step) DropOwnership() error {
	if !o.OwnedByCurrent() {
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
func (o *Ownable2Step) OwnedByCurrent() bool {
	return runtime.CurrentRealm().Address() == o.owner
}

// AssertOwnedByCurrent panics if the caller is not the owner
func (o *Ownable2Step) AssertOwnedByCurrent() {
	if runtime.CurrentRealm().Address() != o.owner {
		panic(ErrUnauthorized)
	}
}

// OwnedByPrevious checks if the caller of the function is the Realm's owner
func (o *Ownable2Step) OwnedByPrevious() bool {
	return runtime.PreviousRealm().Address() == o.owner
}

// AssertOwnedByPrevious panics if the caller is not the owner
func (o *Ownable2Step) AssertOwnedByPrevious() {
	if runtime.PreviousRealm().Address() != o.owner {
		panic(ErrUnauthorized)
	}
}
