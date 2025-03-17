package users

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"

	susers "gno.land/r/sys/users"
)

var (
	alice     = "alice123"
	bob       = "bob123"
	aliceAddr = testutils.TestAddress(alice)
	bobAddr   = testutils.TestAddress(bob)
)

func TestRegister_Valid(t *testing.T) {
	std.TestSetOriginSend(std.NewCoins(std.NewCoin("ugnot", 1_000_000)), nil)
	std.TestSetRealm(std.NewUserRealm(aliceAddr))
	std.TestSetOriginCaller(aliceAddr)

	uassert.NotPanics(t, func() {
		Register(alice)
	})

	res, latest := susers.ResolveName(alice)

	uassert.NotEqual(t, nil, res)
	uassert.Equal(t, alice, res.Name())
	uassert.Equal(t, aliceAddr, res.Addr())
	uassert.False(t, res.IsDeleted())
	uassert.True(t, latest)
}

func TestRegister_Invalid(t *testing.T) {
	std.TestSetOriginSend(std.NewCoins(std.NewCoin("ugnot", 1_000_000)), nil)
	std.TestSetRealm(std.NewUserRealm(bobAddr))
	std.TestSetOriginCaller(bobAddr)

	// Invalid usernames
	uassert.PanicsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register("alice") // vanity
	})

	uassert.PanicsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register("") // empty
	})

	uassert.PanicsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register("    ") // empty
	})

	uassert.PanicsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register("123") // empty
	})

	uassert.PanicsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register("123") // only numbers
	})

	uassert.PanicsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register("alice&#($)") // non-allowed chars
	})

	uassert.PanicsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register("Alice123") // upper-case
	})

	uassert.PanicsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register("toolongusernametoolongusernametoolongusername123") // too long
	})

	// Name taken
	urequire.NotPanics(t, func() {
		Register(bob)
	})

	uassert.PanicsWithMessage(t, susers.ErrNameTaken.Error(), func() {
		Register(bob) // already registered
	})
}
func TestRegister_InvalidPayment(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(bobAddr))
	std.TestSetOriginCaller(bobAddr)

	std.TestSetOriginSend(std.NewCoins(std.NewCoin("ugnot", 12)), nil) // invalid payment amount

	uassert.PanicsWithMessage(t, ErrInvalidPayment.Error(), func() {
		Register(alice)
	})
}
