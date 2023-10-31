package ownable

import (
	"std"
)

// Ownable is meant to be used as a top-level object to make your contract ownable OR
// being embedded in a Gno object to manage per-object ownership.
type Ownable struct {
	owner std.Address
}

func New() *Ownable {
	return &Ownable{
		owner: std.GetOrigCaller(),
	}
}

// TransferOwnership transfers ownership of the Ownable struct to a new address
func (o *Ownable) TransferOwnership(newOwner std.Address) error {
	if !o.CallerIsOwner() {
		return ErrUnauthorized
	}

	if !newOwner.IsValid() {
		return ErrInvalidAddress
	}

	o.owner = newOwner
	return nil
}

// DropOwnership Removes the owner, effectively disabling any owner-related actions
// Top-level usage: disables all only-owner actions/functions,
// Embedded usage: behaves like a burn functionality, removing the owner from the struct
func (o *Ownable) DropOwnership() error {
	if !o.CallerIsOwner() {
		return ErrUnauthorized
	}

	o.owner = ""
	return nil
}

// CallerIsOwner checks if the caller of the function is the Realm's owner
func (o *Ownable) CallerIsOwner() bool {
	if std.GetOrigCaller() == o.owner {
		return true
	}
	return false
}

func (o *Ownable) GetOwner() std.Address {
	return o.owner
}
