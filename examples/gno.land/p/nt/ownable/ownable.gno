package ownable

import (
	"chain"
	"chain/runtime"
)

const OwnershipTransferEvent = "OwnershipTransfer"

// Ownable is meant to be used as a top-level object to make your contract ownable OR
// being embedded in a Gno object to manage per-object ownership.
// Ownable is safe to export as a top-level object
type Ownable struct {
	owner address
}

func New() *Ownable {
	return &Ownable{
		owner: runtime.CurrentRealm().Address(),
	}
}

func NewWithOrigin() *Ownable {
	origin := runtime.OriginCaller()
	previous := runtime.PreviousRealm()
	if origin != previous.Address() {
		panic("NewWithOrigin() should be called from init() where std.PreviousRealm() is origin")
	}
	return &Ownable{
		owner: origin,
	}
}

func NewWithAddress(addr address) *Ownable {
	return &Ownable{
		owner: addr,
	}
}

// TransferOwnership transfers ownership of the Ownable to a new address.
// It checks that runtime.CurrentRealm() matches the current owner.
func (o *Ownable) TransferOwnership(newOwner address) error {
	if !o.OwnedByCurrent() {
		return ErrUnauthorized
	}
	return o.transferOwnership(newOwner)
}

// TransferOwnershipByPrevious transfers ownership of the Ownable to a new address.
// It checks that runtime.PreviousRealm() matches the current owner.
// This is symmetric with DropOwnershipByPrevious and has the same safety
// profile: only the actual owner (as PreviousRealm) can succeed.
func (o *Ownable) TransferOwnershipByPrevious(newOwner address) error {
	if !o.OwnedByPrevious() {
		return ErrUnauthorized
	}
	return o.transferOwnership(newOwner)
}

func (o *Ownable) transferOwnership(newOwner address) error {
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

// DropOwnershipByCurrent removes the owner, effectively disabling any owner-related actions
// Top-level usage: disables all only-owner actions/functions,
// Embedded usage: behaves like a burn functionality, removing the owner from the struct
func (o *Ownable) DropOwnershipByCurrent() error {
	if !o.OwnedByCurrent() {
		return ErrUnauthorized
	}
	o.dropOwnership(o.owner)
	return nil
}

// DropOwnershipByPrevious removes the owner, effectively disabling any owner-related actions
// Top-level usage: disables all only-owner actions/functions,
// Embedded usage: behaves like a burn functionality, removing the owner from the struct
func (o *Ownable) DropOwnershipByPrevious() error {
	if !o.OwnedByPrevious() {
		return ErrUnauthorized
	}
	o.dropOwnership(o.owner)
	return nil
}

func (o *Ownable) dropOwnership(prevOwner address) {
	o.owner = ""
	chain.Emit(
		OwnershipTransferEvent,
		"from", prevOwner.String(),
		"to", "",
	)
}

// Owner returns the owner address from Ownable
func (o *Ownable) Owner() address {
	if o == nil {
		return address("")
	}
	return o.owner
}

// OwnedByCurrent checks if the caller of the function is the Realm's owner
func (o *Ownable) OwnedByCurrent() bool {
	if o == nil {
		return false
	}
	return runtime.CurrentRealm().Address() == o.owner
}

// AssertOwnedByCurrent panics if the caller is not the owner
func (o *Ownable) AssertOwnedByCurrent() {
	if !o.OwnedByCurrent() {
		panic(ErrUnauthorized)
	}
}

// OwnedByPrevious checks if the caller of the function is the Realm's owner
func (o *Ownable) OwnedByPrevious() bool {
	if o == nil {
		return false
	}
	return runtime.PreviousRealm().Address() == o.owner
}

// AssertOwnedByPrevious panics if the caller is not the owner
func (o *Ownable) AssertOwnedByPrevious() {
	if !o.OwnedByPrevious() {
		panic(ErrUnauthorized)
	}
}
