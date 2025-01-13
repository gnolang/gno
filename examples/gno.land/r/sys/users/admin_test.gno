package users

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestAddToWhitelist(t *testing.T) {
	t.Run("invalid new address", func(t *testing.T) {
		std.TestSetRealm(adminRealm)

		uassert.Error(t, AddToWhitelist(std.Address("invalidaddress")), ErrInvalidAddress.Error())
		uassert.Error(t, AddToWhitelist(std.Address("")), ErrInvalidAddress.Error())
		uassert.Error(t, AddToWhitelist(std.Address("000000000000000000000000000000000000000000000000")), ErrInvalidAddress.Error())
	})

	t.Run("already whitelisted", func(t *testing.T) {
		std.TestSetRealm(adminRealm)

		uassert.Error(t, AddToWhitelist(whitelistedCallerAddr), ErrAlreadyWhitelisted.Error())
	})

	t.Run("valid addition", func(t *testing.T) {
		std.TestSetRealm(adminRealm)

		uassert.NoError(t, AddToWhitelist(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")))
	})
}

func TestDelFromWhitelist(t *testing.T) {
	t.Run("not on whitelist", func(t *testing.T) {
		std.TestSetRealm(adminRealm)

		uassert.Error(t, DelFromWhitelist(std.Address("")), ErrNotWhitelisted.Error())
	})

	t.Run("valid deletion", func(t *testing.T) {
		std.TestSetRealm(adminRealm)

		uassert.NoError(t, DelFromWhitelist(whitelistedCallerAddr))
	})

	// FIXME: should not be needed. realm state should reset after
	// 	each test, and init should happen before running a new one.
	callerWhitelist.Add(std.DerivePkgAddr(gnolandUsers))
}
