package users

import (
	"std"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/pausable"
	"gno.land/p/moul/addrset"
)

const (
	adminAddr    = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq" // @moul
	gnolandUsers = "gno.land/r/gnoland/users"                 // preregistered with store write perms
)

var (
	Ownable  = ownable.NewWithAddress(adminAddr) // Until a DAO is formed to manage this contract
	Pausable = pausable.NewFromOwnable(Ownable)

	callerWhitelist = addrset.Set{}
)

func init() {
	// Set initial whitelisted caller
	callerWhitelist.Add(std.DerivePkgAddr(gnolandUsers))

}

// AddToWhitelist adds a new address to the caller whitelist
func AddToWhitelist(addr std.Address) error {
	if !Ownable.CallerIsOwner() {
		return ownable.ErrUnauthorized
	}

	if !addr.IsValid() {
		return ErrInvalidAddress
	}

	if callerWhitelist.Has(addr) {
		return ErrAlreadyWhitelisted
	}

	callerWhitelist.Add(addr)

	return nil
}

// DelFromWhitelist removes a caller from the whitelist
func DelFromWhitelist(addr std.Address) error {
	if !Ownable.CallerIsOwner() {
		return ownable.ErrUnauthorized
	}

	if !callerWhitelist.Has(addr) {
		return ErrNotWhitelisted
	}

	if ok := callerWhitelist.Remove(addr); !ok {
		panic("failed to remove address from whitelist")
	}

	return nil
}

// IsOnWhitelist checks if the given address has
// permission to write to the user store
func IsOnWhitelist(addr std.Address) bool {
	return callerWhitelist.Has(addr)
}
