package ownable

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/ufmt"
	"std"
)

type Ownable struct {
	owner std.Address
}

// CallerIsOwner checks if the caller of the function is the Realm's owner
func (o *Ownable) CallerIsOwner() bool {
	if std.GetOrigCaller() == o.owner {
		return true
	}
	return false
}

// TransferOwnership transfers ownership to a new address
// It also checks that the new owner address conforms with the bech32 standard
func (o *Ownable) TransferOwnership(newOwner std.Address) string {
	if !o.CallerIsOwner() {
		return "unauthorized; caller is not owner"
	}

	if !newOwner.Valid() {
		return "new owner address is not valid"
	}

	o.setOwner(newOwner)
	return ufmt.Sprintf("owner changed to %s\n", string(newOwner))
}

// RenounceOwnership Removes the owner from the Realm,
// Warning - Calling RenounceOwnership practically bricks only-owner functions
func (o *Ownable) RenounceOwnership() string {
	if !o.CallerIsOwner() {
		return "unauthorized; caller is not owner"
	}

	o.owner = ""
	return "renounced ownership successfully"
}

func (o *Ownable) GetOwner() std.Address {
	return o.owner
}

func (o *Ownable) setOwner(newOwner std.Address) {
	o.owner = newOwner
}
