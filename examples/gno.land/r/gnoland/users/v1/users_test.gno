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

	uassert.NoError(t, Register(alice))
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
	uassert.Error(t, Register("alice"), ErrInvalidUsername.Error())      // vanity
	uassert.Error(t, Register(""), ErrInvalidUsername.Error())           // empty
	uassert.Error(t, Register("   "), ErrInvalidUsername.Error())        // empty
	uassert.Error(t, Register("123"), ErrInvalidUsername.Error())        // only numbers
	uassert.Error(t, Register("alice&#($)"), ErrInvalidUsername.Error()) // non-allowed chars
	uassert.Error(t, Register("Alice123"), ErrInvalidUsername.Error())   // upper-case
	uassert.Error(t, Register("toolongusernametoolongusernametoolongusername123"),
		ErrInvalidUsername.Error()) // too long

	// Name taken
	urequire.NoError(t, Register(bob))
	uassert.Error(t, Register(bob), susers.ErrNameTaken.Error())
}

func TestRegister_InvalidPayment(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(bobAddr))
	std.TestSetOriginCaller(bobAddr)

	std.TestSetOriginSend(std.NewCoins(std.NewCoin("ugnot", 12)), nil) // invalid payment amount
	uassert.Error(t, Register("alice"), ErrInvalidPayment.Error())     // vanity
}
