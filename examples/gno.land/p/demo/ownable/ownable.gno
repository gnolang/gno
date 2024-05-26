package ownable

import (
	"std"
)

// Ownable defines an interface that requires a method to get the owner's address
type Ownable interface {
	Owner() std.Address
}

// Transferrable embeds Ownable, indicating that an object can transfer ownership
type Transferrable interface {
	Ownable

	TransferOwnership(newOwner std.Address) error
}

// DefaultOwnable provides a basic implementation of ownership functionality
// It can be used as a top-level object to make a contract ownable or embedded within
// another object to manage per-object ownership.
type DefaultOwnable struct {
	owner std.Address
}

func New() *DefaultOwnable {
	return &DefaultOwnable{
		owner: std.GetOrigCaller(),
	}
}

// TransferOwnership transfers ownership of the Ownable struct to a new address
func (o *DefaultOwnable) TransferOwnership(newOwner std.Address) error {
	err := o.CallerIsOwner()
	if err != nil {
		return err
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
func (o *DefaultOwnable) DropOwnership() error {
	err := o.CallerIsOwner()
	if err != nil {
		return err
	}

	o.owner = ""
	return nil
}

// CallerIsOwner checks if the caller of the function is the Realm's owner
func (o *DefaultOwnable) CallerIsOwner() error {
	if std.GetOrigCaller() == o.owner {
		return nil
	}
	return ErrUnauthorized
}

func (o *DefaultOwnable) Owner() std.Address {
	return o.owner
}
