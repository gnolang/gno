package ownable

import (
	"chain"
	"chain/runtime"
)

const OwnershipTransferEvent = "OwnershipTransfer"

// Ownable is meant to be used as a top-level object to make your contract ownable OR
// being embedded in a Gno object to manage per-object ownership.
// Ownable is safe to export as a top-level object.
//
// The auth mode (current-realm vs previous-realm) is decided at construction time:
//   - New() and NewWithAddress() use current-realm auth.
//   - NewWithOrigin() and NewWithAddressByPrevious() use previous-realm auth.
//
// All ownership operations (TransferOwnership, DropOwnership, Owned, AssertOwned)
// automatically use the configured auth mode.
type Ownable struct {
	owner    address
	previous bool // if true, ownership is checked against the previous realm
}

// New creates an Ownable owned by the current realm, using current-realm auth.
func New() *Ownable {
	return &Ownable{
		owner: runtime.CurrentRealm().Address(),
	}
}

// NewWithOrigin creates an Ownable owned by the origin caller (the user who
// initiated the transaction). Must be called from init() where
// PreviousRealm() is the origin caller. Uses previous-realm auth.
func NewWithOrigin() *Ownable {
	origin := runtime.OriginCaller()
	previous := runtime.PreviousRealm()
	if origin != previous.Address() {
		panic("NewWithOrigin() should be called from init() where std.PreviousRealm() is origin")
	}
	return &Ownable{
		owner:    origin,
		previous: true,
	}
}

// NewWithAddress creates an Ownable with the given address as owner,
// using current-realm auth.
func NewWithAddress(addr address) *Ownable {
	return &Ownable{
		owner: addr,
	}
}

// NewWithAddressByPrevious creates an Ownable with the given address as owner,
// using previous-realm auth.
func NewWithAddressByPrevious(addr address) *Ownable {
	return &Ownable{
		owner:    addr,
		previous: true,
	}
}

// Owned returns true if the caller is the owner, according to the configured auth mode.
func (o *Ownable) Owned() bool {
	if o == nil {
		return false
	}
	if o.previous {
		return runtime.PreviousRealm().Address() == o.owner
	}
	return runtime.CurrentRealm().Address() == o.owner
}

// AssertOwned panics if the caller is not the owner, according to the configured auth mode.
func (o *Ownable) AssertOwned() {
	if !o.Owned() {
		panic(ErrUnauthorized)
	}
}

// TransferOwnership transfers ownership of the Ownable to a new address.
// Uses the configured auth mode to verify the caller.
func (o *Ownable) TransferOwnership(newOwner address) error {
	if !o.Owned() {
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

// DropOwnership removes the owner, effectively disabling any owner-related actions.
// Uses the configured auth mode to verify the caller.
func (o *Ownable) DropOwnership() error {
	if !o.Owned() {
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

// Owner returns the owner address from Ownable.
func (o *Ownable) Owner() address {
	if o == nil {
		return address("")
	}
	return o.owner
}
