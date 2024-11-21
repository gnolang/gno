package ownable

import "std"

const OwnershipTransferEvent = "OwnershipTransfer"

// Ownable is meant to be used as a top-level object to make your contract ownable OR
// being embedded in a Gno object to manage per-object ownership.
// Do not export this object, as its methods do not validate the caller.
// For this, check out SafeOwnable.
type Ownable struct {
	owner std.Address
}

func New() *Ownable {
	return &Ownable{
		owner: std.PrevRealm().Addr(),
	}
}

func NewWithAddress(addr std.Address) *Ownable {
	return &Ownable{
		owner: addr,
	}
}

// TransferOwnership transfers ownership of the Ownable struct to a new address
func (o *Ownable) TransferOwnership(newOwner std.Address) error {
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
func (o *Ownable) DropOwnership() {
	prevOwner := o.owner
	o.owner = ""

	std.Emit(
		OwnershipTransferEvent,
		"from", prevOwner.String(),
		"to", "",
	)
}

// Owner returns the owner address from Ownable
func (o Ownable) Owner() std.Address {
	return o.owner
}

// CallerIsOwner checks if the caller of the function is the Realm's owner
func (o Ownable) CallerIsOwner() bool {
	return std.PrevRealm().Addr() == o.owner
}

// AssertCallerIsOwner panics if the caller is not the owner
func (o Ownable) AssertCallerIsOwner() {
	if std.PrevRealm().Addr() != o.owner {
		panic(ErrUnauthorized)
	}
}

// Safe-object

// SafeOwnable is an exportable struct that wraps Ownable
// By exposing a SafeOwnable object, a realm will expose all SafeOwnable exported methods
// to the public automatically. These methods, as per the implementation,
// are only callable by the Ownable's admin.
type SafeOwnable struct {
	ownable *Ownable
}

func (o *Ownable) NewSafeOwnable() *SafeOwnable {
	return &SafeOwnable{
		ownable: o,
	}
}

func (so *SafeOwnable) TransferOwnership(newOwner std.Address) error {
	if !so.ownable.CallerIsOwner() {
		return ErrUnauthorized
	}

	if err := so.ownable.TransferOwnership(newOwner); err != nil {
		return err
	}

	return nil
}

func (so *SafeOwnable) DropOwnership() error {
	if !so.ownable.CallerIsOwner() {
		return ErrUnauthorized
	}

	so.ownable.DropOwnership()
	return nil
}
