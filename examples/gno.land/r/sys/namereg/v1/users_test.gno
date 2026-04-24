package users

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	susers "gno.land/r/sys/users"
)

func init() {
	// Unit tests run with DefaultHeight=123, so the production init() in init.gno
	// (guarded on height==0) is a no-op in tests. Temporarily reset height so we
	// can whitelist this realm as a controller for testing.
	testing.SetHeight(0)
	susers.AddControllerAtGenesis(cross, chain.PackageAddress("gno.land/r/sys/namereg/v1"))
	testing.SetHeight(123)
}

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

		// Substring-match because errInvalidPayment() formats with the
		// current registerPrice (audit #13) — exact-string matching
		// would couple this test to the price literal.
		uassert.AbortsContains(t, ErrInvalidPayment.Error(), func() {
			Register(cross, "alice123")
		})
	})

	// Audit finding #11: the OriginSend() payment check is only trustworthy
	// when the direct caller is a pure EOA (IsUserCall). Any intermediate
	// code realm can attach -send to the tx, keep the coins, and call
	// Register; OriginSend() would still describe the envelope but namereg
	// would receive nothing. The EOA-only guard must reject all such callers.
	t.Run("intermediate code realm is rejected", func(t *testing.T) {
		// PreviousRealm is a published code realm (e.g. a wrapper contract
		// trying to skim registration fees). IsUserCall() must be false.
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/evil/wrapper"))
		testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))

		uassert.AbortsWithMessage(t, ErrNonUserCall.Error(), func() {
			Register(cross, "alice123")
		})
	})

	// Audit finding #13: the OriginSend mismatch error must report the
	// CURRENT registerPrice, not the value frozen at package init time.
	// Before the fix, ErrInvalidPayment was constructed once with
	// ufmt.Errorf at init and silently drifted out of date after every
	// ProposeNewRegisterPrice execution. Now errInvalidPayment() formats
	// lazily.
	t.Run("payment error reflects current registerPrice", func(t *testing.T) {
		// Save and restore registerPrice so subsequent tests in this file
		// see the default 1 GNOT.
		oldPrice := registerPrice
		defer func() { registerPrice = oldPrice }()

		// Set a non-default price the realm wasn't initialized with.
		registerPrice = 5_000_000

		addr := testutils.TestAddress("payer")
		testing.SetRealm(testing.NewUserRealm(addr))
		testing.SetOriginCaller(addr)
		// Send the OLD default amount, which is now wrong.
		testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", 1_000_000)))

		// The panic message must include "5000000" (the current price),
		// not "1000000" (the value frozen at init in the old code).
		uassert.AbortsContains(t, "5000000", func() {
			Register(cross, "abc999")
		})

		// Sanity: still also includes the sentinel stem so callers using
		// substring matching against ErrInvalidPayment continue to work.
		uassert.AbortsContains(t, ErrInvalidPayment.Error(), func() {
			Register(cross, "abc999")
		})
	})
}

// Audit finding #14: ProposeNewRegisterPrice originally rejected only
// negative prices, accepting zero as a valid value. A passed proposal
// to set price=0 would silently make registration free, removing the
// only economic anti-squatting deterrent. The fix enforces a minimum
// of MinRegisterPrice (0.1 GNOT) at proposal-creation time.
func TestProposeNewRegisterPrice_floor(t *testing.T) {
	t.Run("zero is rejected", func(t *testing.T) {
		urequire.PanicsWithMessage(t,
			"price below floor: 0 ugnot < 100000 ugnot (MinRegisterPrice)",
			func() { ProposeNewRegisterPrice(0) })
	})

	t.Run("negative is rejected", func(t *testing.T) {
		urequire.PanicsWithMessage(t,
			"price below floor: -1 ugnot < 100000 ugnot (MinRegisterPrice)",
			func() { ProposeNewRegisterPrice(-1) })
	})

	t.Run("just below floor is rejected", func(t *testing.T) {
		urequire.PanicsWithMessage(t,
			"price below floor: 99999 ugnot < 100000 ugnot (MinRegisterPrice)",
			func() { ProposeNewRegisterPrice(MinRegisterPrice - 1) })
	})

	t.Run("at floor is accepted", func(t *testing.T) {
		// MinRegisterPrice itself must pass — the floor is inclusive.
		urequire.NotPanics(t, func() { ProposeNewRegisterPrice(MinRegisterPrice) })
	})

	t.Run("above floor is accepted", func(t *testing.T) {
		urequire.NotPanics(t, func() { ProposeNewRegisterPrice(1_000_000_000) }) // 1k GNOT
	})
}
