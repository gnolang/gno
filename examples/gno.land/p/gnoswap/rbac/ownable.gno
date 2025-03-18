package rbac

// TODO: copy from examples/gno.land/p/oxtekgrinder/ownable2step package.
// After the audit, we need to modify the package to use it directly.

import (
	"errors"
	"std"
)

var (
	ErrNoPendingOwner      = errors.New("no pending owner")
	ErrUnauthorized        = errors.New("caller is not owner")
	ErrPendingUnauthorized = errors.New("caller is not pending owner")
	ErrInvalidAddress      = errors.New("new owner address is invalid")
)

const OwnershipTransferEvent = "OwnershipTransfer"

// Ownable2Step is a two-step ownership transfer package
// It allows the current owner to set a new owner and the new owner will need to accept the ownership before it is transferred
type Ownable2Step struct {
	owner        std.Address
	pendingOwner std.Address
}

func NewOwnable2Step() *Ownable2Step {
	return &Ownable2Step{
		owner:        std.PreviousRealm().Address(),
		pendingOwner: "",
	}
}

func NewOwnable2StepWithAddress(addr std.Address) *Ownable2Step {
	return &Ownable2Step{
		owner:        addr,
		pendingOwner: "",
	}
}

// TransferOwnership initiate the transfer of the ownership to a new address by setting the PendingOwner
func (o *Ownable2Step) TransferOwnership(newOwner std.Address) error {
	if !o.CallerIsOwner() {
		return ErrUnauthorized
	}
	if !newOwner.IsValid() {
		return ErrInvalidAddress
	}

	o.pendingOwner = newOwner
	return nil
}

// AcceptOwnership accepts the pending ownership transfer
func (o *Ownable2Step) AcceptOwnership() error {
	if o.pendingOwner.String() == "" {
		return ErrNoPendingOwner
	}
	if std.PreviousRealm().Address() != o.pendingOwner {
		return ErrPendingUnauthorized
	}

	o.owner = o.pendingOwner
	o.pendingOwner = ""

	return nil
}

// DropOwnership removes the owner, effectively disabling any owner-related actions
// Top-level usage: disables all only-owner actions/functions,
// Embedded usage: behaves like a burn functionality, removing the owner from the struct
func (o *Ownable2Step) DropOwnership() error {
	if !o.CallerIsOwner() {
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
func (o *Ownable2Step) Owner() std.Address {
	return o.owner
}

// PendingOwner returns the pending owner address from Ownable2Step
func (o *Ownable2Step) PendingOwner() std.Address {
	return o.pendingOwner
}

// CallerIsOwner checks if the caller of the function is the Realm's owner
func (o *Ownable2Step) CallerIsOwner() bool {
	return std.PreviousRealm().Address() == o.owner
}

// AssertCallerIsOwner panics if the caller is not the owner
func (o *Ownable2Step) AssertCallerIsOwner() {
	if std.PreviousRealm().Address() != o.owner {
		panic(ErrUnauthorized)
	}
}
