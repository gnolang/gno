package users

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"

	susers "gno.land/r/sys/users"
)

func TestRegister_Valid(t *testing.T) {
	validUsernames := []string{
		"abc",                  // 3 letters min
		"abcd",                 // letters-only (vanity)
		"abc123",               // letters + digits
		"abc_def",              // letters + underscore
		"abc123_def456",        // mix of letters / digits / underscore
		"abc_defghi_jklmnopq",  // 19 chars
		"abc_defghi_jklmnopqr", // 20 chars (max)
	}

	for _, username := range validUsernames {
		addr := testutils.TestAddress(username)

		// Simulate a proper user call with the required 1GNOT fee.
		testing.SetRealm(std.NewUserRealm(addr))
		testing.SetOriginCaller(addr)
		testing.SetOriginSend(std.NewCoins(std.NewCoin("ugnot", registerPrice)))

		urequire.NotPanics(t, func() {
			Register(cross, username)
		})
	}
}

func TestRegister_Invalid(t *testing.T) {
	testing.SetOriginSend(std.NewCoins(std.NewCoin("ugnot", registerPrice)))
	testing.SetRealm(std.NewUserRealm(testutils.TestAddress("bob123")))

	// Invalid usernames
	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register(cross, "") // empty
	})

	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register(cross, "    ") // empty
	})

	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register(cross, "123") // empty
	})

	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register(cross, "123") // only numbers
	})

	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register(cross, "alice&#($)") // non-allowed chars
	})

	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register(cross, "Alice123") // upper-case
	})

	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register(cross, "abcdefghijklmnopqrstu") // 21 chars
	})

	uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
		Register(cross, "toolongusernametoolongusernametoolongusername123") // too long
	})

	username := "bob123"

	// Name taken
	urequire.NotPanics(t, func() {
		Register(cross, username)
	})

	uassert.AbortsWithMessage(t, susers.ErrNameTaken.Error(), func() {
		Register(cross, username) // already registered
	})
}

func TestRegister_InvalidPayment(t *testing.T) {
	addr := testutils.TestAddress("bob123")

	testing.SetRealm(std.NewUserRealm(addr))
	testing.SetOriginCaller(addr)

	testing.SetOriginSend(std.NewCoins(std.NewCoin("ugnot", 12))) // invalid payment amount

	uassert.AbortsWithMessage(t, ErrInvalidPayment.Error(), func() {
		Register(cross, "alice123")
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
