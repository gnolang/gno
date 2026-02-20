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
	owner    address
	authMode authMode
}

// authMode defines which realm context is allowed to perform ownership checks.
// authModeLegacy is kept for backward compatibility with previously persisted
// instances that predate mode locking.
type authMode uint8

const (
	authModeLegacy authMode = iota
	authModeCurrent
	authModePrevious
)

func New() *Ownable {
	return &Ownable{
		owner:    runtime.CurrentRealm().Address(),
		authMode: authModeCurrent,
	}
}

func NewWithOrigin() *Ownable {
	origin := runtime.OriginCaller()
	previous := runtime.PreviousRealm()
	if origin != previous.Address() {
		panic("NewWithOrigin() should be called from init() where std.PreviousRealm() is origin")
	}
	return &Ownable{
		owner:    origin,
		authMode: authModePrevious,
	}
}

// NewWithAddress keeps current-realm auth as default.
func NewWithAddress(addr address) *Ownable {
	return &Ownable{
		owner:    addr,
		authMode: authModeCurrent,
	}
}

// NewWithAddressByPrevious creates an Ownable bound to previous-realm auth.
func NewWithAddressByPrevious(addr address) *Ownable {
	return &Ownable{
		owner:    addr,
		authMode: authModePrevious,
	}
}

// TransferOwnership transfers ownership of the Ownable struct to a new address
func (o *Ownable) TransferOwnership(newOwner address) error {
	if !o.OwnedByCurrent() {
		return ErrUnauthorized
	}

	return o.transferOwnership(newOwner)
}

// TransferOwnershipByPrevious transfers ownership of the Ownable struct
// when ownership is asserted against the previous realm.
func (o *Ownable) TransferOwnershipByPrevious(newOwner address) error {
	if !o.OwnedByPrevious() {
		return ErrUnauthorized
	}

	return o.transferOwnership(newOwner)
}

func (o *Ownable) transferOwnership(newOwner address) error {
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

// DropOwnershipByCurrent removes the owner, effectively disabling any owner-related actions
// Top-level usage: disables all only-owner actions/functions,
// Embedded usage: behaves like a burn functionality, removing the owner from the struct
func (o *Ownable) DropOwnershipByCurrent() error {
	if !o.OwnedByCurrent() {
		return ErrUnauthorized
	}
	o.dropOwnership(o.owner)
	return nil
}

// DropOwnershipByPrevious removes the owner, effectively disabling any owner-related actions
// Top-level usage: disables all only-owner actions/functions,
// Embedded usage: behaves like a burn functionality, removing the owner from the struct
func (o *Ownable) DropOwnershipByPrevious() error {
	if !o.OwnedByPrevious() {
		return ErrUnauthorized
	}
	o.dropOwnership(o.owner)
	return nil
}

func (o *Ownable) dropOwnership(prevOwner address) {
	o.owner = ""
	chain.Emit(
		OwnershipTransferEvent,
		"from", prevOwner.String(),
		"to", "",
	)
}

// Owner returns the owner address from Ownable
func (o *Ownable) Owner() address {
	if o == nil {
		return address("")
	}
	return o.owner
}

// Owned checks ownership according to the Ownable auth mode.
func (o *Ownable) Owned() bool {
	if o == nil {
		return false
	}

	switch o.authMode {
	case authModeCurrent:
		return runtime.CurrentRealm().Address() == o.owner
	case authModePrevious:
		return o.ownedByPreviousRealm()
	default:
		// legacy mode keeps compatibility with pre-mode-lock behavior.
		return runtime.CurrentRealm().Address() == o.owner || o.ownedByPreviousRealm()
	}
}

// AssertOwned panics if the caller is not the owner according to auth mode.
func (o *Ownable) AssertOwned() {
	if !o.Owned() {
		panic(ErrUnauthorized)
	}
}

// OwnedByCurrent checks if the caller of the function is the Realm's owner
func (o *Ownable) OwnedByCurrent() bool {
	if o == nil {
		return false
	}
	if !o.allowsCurrentRealmAuth() {
		return false
	}
	return runtime.CurrentRealm().Address() == o.owner
}

// AssertOwnedByCurrent panics if the caller is not the owner
func (o *Ownable) AssertOwnedByCurrent() {
	if !o.OwnedByCurrent() {
		panic(ErrUnauthorized)
	}
}

// OwnedByPrevious checks if the caller of the function is the Realm's owner
func (o *Ownable) OwnedByPrevious() bool {
	if o == nil {
		return false
	}
	if !o.allowsPreviousRealmAuth() {
		return false
	}
	return o.ownedByPreviousRealm()
}

// AssertOwnedByPrevious panics if the caller is not the owner
func (o *Ownable) AssertOwnedByPrevious() {
	if !o.OwnedByPrevious() {
		panic(ErrUnauthorized)
	}
}

func (o *Ownable) allowsCurrentRealmAuth() bool {
	return o.authMode != authModePrevious
}

func (o *Ownable) allowsPreviousRealmAuth() bool {
	return o.authMode != authModeCurrent
}

func (o *Ownable) ownedByPreviousRealm() (owned bool) {
	defer func() {
		if recover() != nil {
			owned = false
		}
	}()

	return runtime.PreviousRealm().Address() == o.owner
}
