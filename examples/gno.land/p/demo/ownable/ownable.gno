package ownable

import (
	"std"
)

const OwnershipTransferEvent = "OwnershipTransfer"

// Ownable is meant to be used as a top-level object to make your contract ownable OR
// being embedded in a Gno object to manage per-object ownership.
type Ownable interface {
	Owner() std.Address
}

// Transferrable embeds Ownable, indicating that the owner can transfer ownership and drop ownership
type Transferrable interface {
	Ownable

	TransferOwnership(newOwner std.Address) error
	DropOwnership() error
}

// transferrable provides a basic implementation of ownership functionality
// It can be used as a top-level object to make a contract ownable or embedded within
// another object to manage per-object ownership.
type transferrable struct {
	owner std.Address
}

// Ensure that transferrable implements the Ownable and Transferrable interfaces
var (
	_ Ownable       = (*transferrable)(nil)
	_ Transferrable = (*transferrable)(nil)
)

func New() *transferrable {
	return &transferrable{
		owner: std.PrevRealm().Addr(),
	}
}

func NewWithAddress(addr std.Address) *transferrable {
	return &transferrable{
		owner: addr,
	}
}

// TransferOwnership transfers ownership of the Ownable struct to a new address
func (o *transferrable) TransferOwnership(newOwner std.Address) error {
	if err := o.CallerIsOwner(); err != nil {
		return err
	}

	if !newOwner.IsValid() {
		return ErrInvalidAddress
	}

	prevOwner := o.owner
	o.owner = newOwner
	std.Emit(
		OwnershipTransferEvent,
		"from", string(prevOwner),
		"to", string(newOwner),
	)
	return nil
}

// DropOwnership removes the owner, effectively disabling any owner-related actions
// Top-level usage: disables all only-owner actions/functions,
// Embedded usage: behaves like a burn functionality, removing the owner from the struct
func (o *transferrable) DropOwnership() error {
	if err := o.CallerIsOwner(); err != nil {
		return err
	}

	prevOwner := o.owner
	o.owner = ""

	std.Emit(
		OwnershipTransferEvent,
		"from", string(prevOwner),
		"to", "",
	)

	return nil
}

func (o transferrable) Owner() std.Address {
	return o.owner
}

// CallerIsOwner checks if the caller of the function is the Realm's owner
func (o transferrable) CallerIsOwner() error {
	if std.PrevRealm().Addr() == o.owner {
		return nil
	}

	return ErrUnauthorized
}

func (o transferrable) AssertCallerIsOwner() {
	if std.PrevRealm().Addr() != o.owner {
		panic(ErrUnauthorized)
	}
}
