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
	owner chain.Address
}

func New() *Ownable {
	return &Ownable{
		owner: runtime.PreviousRealm().Address(),
	}
}

func NewWithAddress(addr chain.Address) *Ownable {
	return &Ownable{
		owner: addr,
	}
}

// TransferOwnership transfers ownership of the Ownable struct to a new address
func (o *Ownable) TransferOwnership(newOwner chain.Address) error {
	if !o.CallerIsOwner() {
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

// DropOwnership removes the owner, effectively disabling any owner-related actions
// Top-level usage: disables all only-owner actions/functions,
// Embedded usage: behaves like a burn functionality, removing the owner from the struct
func (o *Ownable) DropOwnership() error {
	if !o.CallerIsOwner() {
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

// Owner returns the owner address from Ownable
func (o *Ownable) Owner() chain.Address {
	if o == nil {
		return chain.Address("")
	}
	return o.owner
}

// CallerIsOwner checks if the caller of the function is the Realm's owner
func (o *Ownable) CallerIsOwner() bool {
	if o == nil {
		return false
	}
	return runtime.PreviousRealm().Address() == o.owner
}

// AssertCallerIsOwner panics if the caller is not the owner
func (o *Ownable) AssertCallerIsOwner() {
	if o == nil {
		panic(ErrUnauthorized)
	}
	caller := runtime.PreviousRealm().Address()
	if caller != o.owner {
		panic(ErrUnauthorized)
	}
}
