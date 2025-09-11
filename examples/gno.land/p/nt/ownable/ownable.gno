package ownable

import (
	"std"
)

const OwnershipTransferEvent = "OwnershipTransfer"

// Ownable is meant to be used as a top-level object to make your contract ownable OR
// being embedded in a Gno object to manage per-object ownership.
// Ownable is safe to export as a top-level object
type Ownable struct {
	owner std.Address
}

func New() *Ownable {
	return &Ownable{
		owner: std.CurrentRealm().Address(),
	}
}

func NewWithOrigin() *Ownable {
	origin := std.OriginCaller()
	previous := std.PreviousRealm()
	if origin != previous.Address() {
		panic("NewWithOrigin() should be called from init() where std.PreviousRealm() is origin")
	}
	return &Ownable{
		owner: origin,
	}
}

func NewWithAddress(addr std.Address) *Ownable {
	return &Ownable{
		owner: addr,
	}
}

// TransferOwnership transfers ownership of the Ownable struct to a new address
func (o *Ownable) TransferOwnership(newOwner std.Address) error {
	if !o.OwnedByCurrent() {
		return ErrUnauthorized
	}

	if !newOwner.IsValid() {
		return ErrInvalidAddress
	}

	prevOwner := o.owner
	o.owner = newOwner
	std.Emit(
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

func (o *Ownable) dropOwnership(prevOwner std.Address) {
	o.owner = ""
	std.Emit(
		OwnershipTransferEvent,
		"from", prevOwner.String(),
		"to", "",
	)
}

// Owner returns the owner address from Ownable
func (o *Ownable) Owner() std.Address {
	if o == nil {
		return std.Address("")
	}
	return o.owner
}

// OwnedByCurrent checks if the caller of the function is the Realm's owner
func (o *Ownable) OwnedByCurrent() bool {
	if o == nil {
		return false
	}
	return std.CurrentRealm().Address() == o.owner
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
	return std.PreviousRealm().Address() == o.owner
}

// AssertOwnedByPrevious panics if the caller is not the owner
func (o *Ownable) AssertOwnedByPrevious() {
	if !o.OwnedByPrevious() {
		panic(ErrUnauthorized)
	}
}
