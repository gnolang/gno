package ownable

import (
	"std"

	"gno.land/p/demo/helper"
)

// Ownable defines an interface that requires a method to get the owner's address
type Ownable interface {
	Owner() std.Address
}

// Transferrable embeds Ownable, indicating that an object can transfer and drop an ownership
type Transferrable interface {
	Ownable

	TransferOwnership(newOwner std.Address) error
	DropOwnership() error
}

// TransferrableOwnership provides a basic implementation of ownership functionality
// It can be used as a top-level object to make a contract ownable or embedded within
// another object to manage per-object ownership.
type TransferrableOwnership struct {
	owner std.Address
}

// Ensure that TransferrableOwnership implements the Ownable and Transferrable interfaces
var (
	_ Ownable       = (*TransferrableOwnership)(nil)
	_ Transferrable = (*TransferrableOwnership)(nil)
)

func New() *TransferrableOwnership {
	return &TransferrableOwnership{
		owner: std.GetOrigCaller(),
	}
}

// TransferOwnership transfers ownership of the Ownable struct to a new address
func (o *TransferrableOwnership) TransferOwnership(newOwner std.Address) error {
	if !helper.PrevRealmIsOwner(o.owner) {
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
func (o *TransferrableOwnership) DropOwnership() error {
	if !helper.PrevRealmIsOwner(o.owner) {
		return ErrUnauthorized
	}

	o.owner = ""
	return nil
}

func (o *TransferrableOwnership) Owner() std.Address {
	return o.owner
}
