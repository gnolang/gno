package ownable

import "std"

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

func NewWithAddress(addr std.Address) *Ownable {
	return &Ownable{
		owner: addr,
	}
}

// TransferOwnership transfers ownership of the Ownable struct to a new address
func (o *Ownable) TransferOwnership(newOwner std.Address) error {
	if !o.OwnedByPrevious() {
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

// DropOwnership removes the owner, effectively disabling any owner-related actions
// Top-level usage: disables all only-owner actions/functions,
// Embedded usage: behaves like a burn functionality, removing the owner from the struct
func (o *Ownable) DropOwnership() error {
	if !o.OwnedByPrevious() {
		return ErrUnauthorized
	}

	prevOwner := o.owner
	o.owner = ""

	std.Emit(
		OwnershipTransferEvent,
		"from", prevOwner.String(),
		"to", "",
	)

	return nil
}

// Owner returns the owner address from Ownable
func (o *Ownable) Owner() std.Address {
	if o == nil {
		return std.Address("")
	}
	return o.owner
}

// OwnedByPrevious( checks if the caller of the function is the Realm's owner
func (o *Ownable) OwnedByPrevious() bool {
	if o == nil {
		return false
	}
	return std.CurrentRealm().Address() == o.owner
}

// AssertOwnedByPrevious panics if the caller is not the owner
func (o *Ownable) AssertOwnedByPrevious() {
	if o == nil {
		panic(ErrUnauthorized)
	}
	caller := std.CurrentRealm().Address()
	if caller != o.owner {
		panic(ErrUnauthorized)
	}
}
