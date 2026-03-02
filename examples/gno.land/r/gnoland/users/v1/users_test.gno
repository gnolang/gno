package users

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

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
		testing.SetRealm(testing.NewUserRealm(addr))
		testing.SetOriginCaller(addr)
		testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))

		urequire.NotPanics(t, func() {
			Register(cross, username)
		})
	}
}

func TestRegister_Free(t *testing.T) {
	registerPrice = 0

	addr := testutils.TestAddress("alice456")

	// Simulate a proper user call with 0gnot fee.
	testing.SetRealm(testing.NewUserRealm(addr))
	testing.SetOriginCaller(addr)

	urequire.NotPanics(t, func() {
		Register(cross, "alice456")
	})
}

func TestRegister_Invalid(t *testing.T) {
	t.Run("invalid usernames", func(t *testing.T) {
		testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
		testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("bob123")))

		invalidUsernames := []string{
			"alice",                 // vanity
			"",                      // empty
			"    ",                  // empty
			"123",                   // only numbers (no letters)
			"abc",                   // only letters (no numbers)
			"alice&#($)123",         // invalid characters
			"Alice123",              // uppercase
			"abcdefghijklmnopqr123", // 21 chars (> max)
			"toolongusernametoolongusernametoolongusername123", // super long
		}

		for _, username := range invalidUsernames {
			uassert.AbortsWithMessage(t, ErrInvalidUsername.Error(), func() {
				Register(cross, username)
			})
		}
	})

	t.Run("taken username", func(t *testing.T) {
		testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
		testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("bob123")))

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

		testing.SetRealm(testing.NewUserRealm(addr))
		testing.SetOriginCaller(addr)

		testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", 12))) // invalid payment amount

		uassert.AbortsWithMessage(t, ErrInvalidPayment.Error(), func() {
			Register(cross, "alice123")
		})
	})
}
