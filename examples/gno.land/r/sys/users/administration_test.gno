package users

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

var (
	whitelistedCaller = std.DerivePkgAddr("gno.land/r/gnoland/users")
	adminRealm        = std.NewUserRealm(adminAddr)
	alice             = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	aliceCodeRealm    = std.NewCodeRealm("gno.land/r/alice/package")
)

func TestIsOnWhitelist(t *testing.T) {
	t.Run("init_caller_registration", func(t *testing.T) {
		t.Parallel()
		uassert.True(t, IsOnWhitelist(whitelistedCaller))
	})

	t.Run("not on whitelist", func(t *testing.T) {
		t.Parallel()
		uassert.False(t, IsOnWhitelist(aliceCodeRealm.Addr()))
	})
}

func TestAddToWhitelist(t *testing.T) {
	t.Run("invalid new address", func(t *testing.T) {
		t.Parallel()
		std.TestSetRealm(adminRealm)

		uassert.Error(t, AddToWhitelist(std.Address("invalidaddress")), ErrInvalidAddress.Error())
		uassert.Error(t, AddToWhitelist(std.Address("")), ErrInvalidAddress.Error())
		uassert.Error(t, AddToWhitelist(std.Address("000000000000000000000000000000000000000000000000")), ErrInvalidAddress.Error())
	})

	t.Run("already whitelisted", func(t *testing.T) {
		t.Parallel()
		std.TestSetRealm(adminRealm)

		uassert.Error(t, AddToWhitelist(whitelistedCaller), ErrAlreadyWhitelisted.Error())
	})

	t.Run("valid addition", func(t *testing.T) {
		t.Parallel()
		std.TestSetRealm(adminRealm)

		uassert.NoError(t, AddToWhitelist("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	})
}

func TestDelFromWhitelist(t *testing.T) {
	t.Run("not on whitelist", func(t *testing.T) {
		t.Parallel()
		std.TestSetRealm(adminRealm)

		uassert.Error(t, DelFromWhitelist(std.Address("")), ErrDoesNotExist.Error())
	})

	t.Run("valid deletion", func(t *testing.T) {
		t.Parallel()
		std.TestSetRealm(adminRealm)

		uassert.NoError(t, DelFromWhitelist(whitelistedCaller))
	})
}
