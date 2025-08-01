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
		"abc123",               // letters + digits
		"abc123_def456",        // mix of letters / digits / underscore
		"abc_defghi_jklmn123",  // 19 chars
		"abc_defghi_jklmno456", // 20 chars (max)
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
	t.Run("invalid usernames", func(t *testing.T) {
		testing.SetOriginSend(std.NewCoins(std.NewCoin("ugnot", registerPrice)))
		testing.SetRealm(std.NewUserRealm(testutils.TestAddress("bob123")))

		invalidUsernames := []string{
			"alice",                 // vanity
			"",                      // empty
			"    ",                  // empty
			"123",                   // only numbers (no letters)
			"abc",                   // only letters (no numbers)
			"alice&#($)123",            // invalid characters
			"Alice123",              // uppercase
			"abcdefghijklmnopqrstu", // 21 chars (> max)
			"toolongusernametoolongusernametoolongusername123", // super long
		}

		for _, username := range invalidUsernames {
			uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
				Register(cross, username)
			})
		}
	})

	t.Run("taken username", func(t *testing.T) {
		testing.SetOriginSend(std.NewCoins(std.NewCoin("ugnot", registerPrice)))
		testing.SetRealm(std.NewUserRealm(testutils.TestAddress("bob123")))

		username := "bob123"

		urequire.NotPanics(t, func() {
			Register(cross, username)
		})

		uassert.AbortsWithMessage(t, susers.ErrNameTaken.Error(), func() {
			Register(cross, username) // already registered
		})
	})

	t.Run("invalid payment", func(t *testing.T) {
		addr := testutils.TestAddress("bob123")

		testing.SetRealm(std.NewUserRealm(addr))
		testing.SetOriginCaller(addr)

		testing.SetOriginSend(std.NewCoins(std.NewCoin("ugnot", 12))) // invalid payment amount

		uassert.AbortsWithMessage(t, ErrInvalidPayment.Error(), func() {
			Register(cross, "alice123")
		})
	})
}
