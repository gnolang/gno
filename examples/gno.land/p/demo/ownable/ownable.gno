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
		owner: std.PrevRealm().Addr(),
	}
}

func NewWithAddress(addr std.Address) *Ownable {
	return &Ownable{owner: addr}
}

// TransferOwnership transfers ownership of the Ownable struct to a new address
func (o *Ownable) TransferOwnership(newOwner std.Address) error {
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
func (o *Ownable) DropOwnership() error {
	err := o.CallerIsOwner()
	if err != nil {
		return err
	}

	o.owner = ""
	return nil
}

func (o Ownable) Owner() std.Address {
	return o.owner
}

// CallerIsOwner checks if the caller of the function is the Realm's owner
func (o Ownable) CallerIsOwner() error {
	if std.PrevRealm().Addr() == o.owner {
		return nil
	}
	return ErrUnauthorized
}

func (o Ownable) AssertCallerIsOwner() {
	if std.PrevRealm().Addr() != o.owner {
		panic(ErrUnauthorized)
	}
}
