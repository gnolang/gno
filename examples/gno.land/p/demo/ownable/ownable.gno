package ownable

import (
	"std"
)

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

// transferrableOwnership provides a basic implementation of ownership functionality
// It can be used as a top-level object to make a contract ownable or embedded within
// another object to manage per-object ownership.
type transferrableOwnership struct {
	owner std.Address
}

// Ensure that transferrableOwnership implements the Ownable and Transferrable interfaces
var (
	_ Ownable       = (*transferrableOwnership)(nil)
	_ Transferrable = (*transferrableOwnership)(nil)
)

func New() *transferrableOwnership {
	return &transferrableOwnership{
		owner: std.PrevRealm().Addr(),
	}
}

// TransferOwnership transfers ownership of the Ownable struct to a new address
func (o *transferrableOwnership) TransferOwnership(newOwner std.Address) error {
	if !callerIsOwner(o.owner) {
		return ErrUnauthorized
	}

	if !newOwner.IsValid() {
		return ErrInvalidAddress
	}

	o.owner = newOwner
	return nil
}

// DropOwnership removes the owner, effectively disabling any owner-related actions
// Top-level usage: disables all only-owner actions/functions,
// Embedded usage: behaves like a burn functionality, removing the owner from the struct
func (o *transferrableOwnership) DropOwnership() error {
	if !callerIsOwner(o.owner) {
		return ErrUnauthorized
	}

	o.owner = ""
	return nil
}

func (o *transferrableOwnership) Owner() std.Address {
	return o.owner
}

func callerIsOwner(addr std.Address) bool {
	return std.PrevRealm().Addr() == addr
}
