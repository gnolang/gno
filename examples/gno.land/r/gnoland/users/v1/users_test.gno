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
	testing.SetOriginSend(std.NewCoins(std.NewCoin("ugnot", 1_000_000)))
	testing.SetRealm(std.NewUserRealm(aliceAddr))

	uassert.NotPanics(t, func() {
		cross(Register)(alice)
	})

	res, latest := susers.ResolveName(alice)
	uassert.True(t, latest)
	uassert.NotTypedNil(t, res)
	if !istypednil(res) {
		uassert.Equal(t, alice, res.Name())
		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.False(t, res.IsDeleted())
	}
}

func TestRegister_Invalid(t *testing.T) {
	testing.SetOriginSend(std.NewCoins(std.NewCoin("ugnot", 1_000_000)))
	testing.SetRealm(std.NewUserRealm(bobAddr))

	// Invalid usernames
	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		cross(Register)("alice") // vanity
	})

	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		cross(Register)("") // empty
	})

	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		cross(Register)("    ") // empty
	})

	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		cross(Register)("123") // empty
	})

	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		cross(Register)("123") // only numbers
	})

	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		cross(Register)("alice&#($)") // non-allowed chars
	})

	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		cross(Register)("Alice123") // upper-case
	})

	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		cross(Register)("toolongusernametoolongusernametoolongusername123") // too long
	})

	// Name taken
	urequire.NotPanics(t, func() {
		cross(Register)(bob)
	})

	uassert.AbortsWithMessage(t, susers.ErrNameTaken.Error(), func() {
		cross(Register)(bob) // already registered
	})
}

func TestRegister_InvalidPayment(t *testing.T) {
	testing.SetRealm(std.NewUserRealm(bobAddr))
	testing.SetOriginCaller(bobAddr)

	testing.SetOriginSend(std.NewCoins(std.NewCoin("ugnot", 12))) // invalid payment amount

	uassert.AbortsWithMessage(t, ErrInvalidPayment.Error(), func() {
		cross(Register)(alice)
	})
}

/* TODO fix this @moul
func TestUsers(t *testing.T) {
	acc1 := std.Address("g1test1")
	fee := "1000ugnot"

	// Check initial state
	uassert.NotPanics(t, func() {
		MustGetByName(susers.DefaultMinFee)
	})
}
*/
