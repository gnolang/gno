package users

import (
	"std"

	"gno.land/p/demo/dao"
	"gno.land/p/moul/addrset"

	"gno.land/r/gov/dao/bridge"
)

const gusersv1 = "gno.land/r/gnoland/users/v1" // preregistered with store write perms

var controllers = addrset.Set{} // caller whitelist

func init() {
	callerWhitelist.Add(std.DerivePkgAddr(gusersv1)) // initially whitelisted
}

// IsOnWhitelist checks if the given address has
// permission to write to the user store
func IsOnWhitelist(addr std.Address) bool {
	return callerWhitelist.Has(addr)
}

// NewAddWhitelistedAddrExecutor allows GovDAO to add a whitelisted caller
func ProposeNewController(addr std.Address) dao.Executor {
	cb := func() error {
		return addToWhitelist(addr)
	}

	return bridge.GovDAO().NewGovDAOExecutor(cb)
}

// NewDeleteWhitelistedAddrExecutor allows GovDAO to add a whitelisted caller
func NewDeleteWhitelistedAddrExecutor(addr std.Address) dao.Executor {
	cb := func() error {
		return deleteFromwhitelist(addr)
	}

	return bridge.GovDAO().NewGovDAOExecutor(cb)
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

	if !callerWhitelist.Add(newCaller) {
		return ErrAlreadyWhitelisted
	}

	return nil
}
