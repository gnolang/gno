package users

import (
	"std"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/pausable"
	"gno.land/p/moul/addrset"
)

const (
	adminAddr    = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq" // @moul > will be GovDAO managed soon, see PR#3523
	gnolandUsers = "gno.land/r/gnoland/users/v1"              // preregistered with store write perms
)

var (
	callerWhitelist = addrset.Set{}

	// Safe objects
	Ownable  = ownable.NewWithAddress(adminAddr)
	Pausable = pausable.NewFromOwnable(Ownable)
)

func init() {
	callerWhitelist.Add(std.DerivePkgAddr(gnolandUsers)) // initially whitelisted
}

// AddToWhitelist adds a new address to the caller whitelist
func AddToWhitelist(addr std.Address) error {
	if !Ownable.CallerIsOwner() {
		return ownable.ErrUnauthorized
	}

	return addToWhitelist(addr)
}

// DelFromWhitelist removes a caller from the whitelist
func DelFromWhitelist(addr std.Address) error {
	if !Ownable.CallerIsOwner() {
		return ownable.ErrUnauthorized
	}

	return deleteFromwhitelist(addr)
}

// IsOnWhitelist checks if the given address has
// permission to write to the user store
func IsOnWhitelist(addr std.Address) bool {
	return callerWhitelist.Has(addr)
}

// Helpers

func deleteFromwhitelist(addr std.Address) error {
	if !callerWhitelist.Has(addr) {
		return ErrNotWhitelisted
	}

	if ok := callerWhitelist.Remove(addr); !ok {
		panic("failed to remove address from whitelist")
	}

	return nil
}

func addToWhitelist(newCaller std.Address) error {
	if !newCaller.IsValid() {
		return ErrInvalidAddress
	}
	if callerWhitelist.Has(newCaller) {
		return ErrAlreadyWhitelisted
	}

	callerWhitelist.Add(newCaller)
	return nil
}
