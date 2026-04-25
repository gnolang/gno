package namereg

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
	// Stems chosen with no `l` characters to keep canonical forms identical
	// to the originals — avoids accidental cross-subtest collisions with
	// other tests that may register l-bearing names.
	validUsernames := []string{
		"nym-bravo123",      // 5-char stem (minimum)
		"nym-tango456",      // 5-char stem
		"nym-victor789",     // 6-char stem
		"nym-romeoecho012",  // 10-char stem
		"nym-mikefoxhote345", // 13-char stem (maximum)
	}

	for _, username := range validUsernames {
		addr := testutils.TestAddress(username)

		testing.SetRealm(testing.NewUserRealm(addr))
		testing.SetOriginCaller(addr)
		testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))

		urequire.NotPanics(t, func() {
			Register(cross, username)
		})
	}
}

func TestRegister_Free(t *testing.T) {
	oldPrice := registerPrice
	defer func() { registerPrice = oldPrice }()
	registerPrice = 0

	addr := testutils.TestAddress("free-payer")

	testing.SetRealm(testing.NewUserRealm(addr))
	testing.SetOriginCaller(addr)

	urequire.NotPanics(t, func() {
		Register(cross, "nym-foxtrot456")
	})
}

func TestRegister_InvalidFormat(t *testing.T) {
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("fmt-payer")))

	cases := []string{
		"",                       // empty
		"    ",                   // whitespace
		"alice123",               // missing nym- prefix
		"usr-alice123",           // wrong prefix
		"nym-",                   // prefix only
		"nym-abc123",             // stem too short (3 chars, need ≥5)
		"nym-abcd123",            // stem too short (4 chars)
		"nym-abcdefghijklmn123",  // stem too long (14 chars, need ≤13)
		"nym-Alice123",           // uppercase in stem
		"nym-al1ce123",           // digit in stem
		"nym-alice12",            // only 2 trailing digits
		"nym-alice1234",          // 4 trailing digits (extra digit in stem too long)
		"nym-alice&#($)123",      // special chars in stem
	}

	for _, username := range cases {
		uassert.AbortsWithMessage(t, ErrInvalidFormat.Error(), func() {
			Register(cross, username)
		})
	}
}

func TestRegister_ReservedPrefix(t *testing.T) {
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("prefix-payer")))

	cases := []string{
		"nym-gnoblah123",   // gno prefix
		"nym-gnomeland456", // gno prefix
		"nym-gillette789",  // gi prefix
		"nym-glasgow012",   // gl prefix (also canonicalizes to gi prefix)
	}

	for _, username := range cases {
		uassert.AbortsWithMessage(t, ErrReservedPrefix.Error(), func() {
			Register(cross, username)
		})
	}
}

func TestRegister_Blacklisted(t *testing.T) {
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("blk-payer")))

	cases := []string{
		// `admin` is in reservedNames; with stem 5 chars it now matches the regex.
		"nym-admin123",
		// `support` is reserved (7 chars).
		"nym-support456",
		// `bitcoin` is reserved (7 chars).
		"nym-bitcoin789",
		// Implicit s-suffix expansion: `blog` is reserved, so `blogs` (stem
		// = blog+s) must also be rejected.
		"nym-blogs012",
		// Canonical match: `setting` is reserved → canonical "setting" (no l).
		// Try `setting` directly via 7-char stem.
		"nym-setting345",
		// Canonical match: `news` is reserved via `new+s` rule; canonical
		// of `news` is `news`. With a 5-char stem this won't reach here
		// (stem must be ≥5). Use a longer reserved name with l→i interaction:
		// `tendermint` (10 chars, no l) → canonical = tendermint → blacklisted.
		"nym-tendermint678",
	}

	for _, username := range cases {
		uassert.AbortsWithMessage(t, ErrBlacklisted.Error(), func() {
			Register(cross, username)
		})
	}
}

func TestRegister_CanonicalCollision(t *testing.T) {
	// First register a name with `l` in the stem.
	addr1 := testutils.TestAddress("collision-1")
	testing.SetRealm(testing.NewUserRealm(addr1))
	testing.SetOriginCaller(addr1)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
	urequire.NotPanics(t, func() {
		Register(cross, "nym-balloon123")
	})

	// Then attempt to register the canonical-equivalent variant
	// (l→i applied: `balloon` → `baiioon`). Must reject with collision.
	addr2 := testutils.TestAddress("collision-2")
	testing.SetRealm(testing.NewUserRealm(addr2))
	testing.SetOriginCaller(addr2)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
	uassert.AbortsWithMessage(t, ErrCanonicalCollision.Error(), func() {
		Register(cross, "nym-baiioon456")
	})

	// Reverse direction also blocked: register a non-l name first, then
	// attempt an l-variant that canonicalizes to the same form.
	addr3 := testutils.TestAddress("collision-3")
	testing.SetRealm(testing.NewUserRealm(addr3))
	testing.SetOriginCaller(addr3)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
	urequire.NotPanics(t, func() {
		Register(cross, "nym-papaiio789")
	})

	addr4 := testutils.TestAddress("collision-4")
	testing.SetRealm(testing.NewUserRealm(addr4))
	testing.SetOriginCaller(addr4)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
	uassert.AbortsWithMessage(t, ErrCanonicalCollision.Error(), func() {
		Register(cross, "nym-papallo012")
	})
}

func TestRegister_TakenUsername(t *testing.T) {
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("taken-payer")))

	username := "nym-zulutwo567"

	urequire.NotPanics(t, func() {
		Register(cross, username)
	})

	// Re-registration of the same exact name is rejected at the canonical-
	// store layer (same canonical → collision) before reaching r/sys/users.
	// (Pre-bridge this would have been rejected by susers.ErrNameTaken; now
	// the namereg-side collision check fires first, which is fine.)
	uassert.AbortsWithMessage(t, ErrCanonicalCollision.Error(), func() {
		Register(cross, username)
	})
}

func TestRegister_InvalidPayment(t *testing.T) {
	addr := testutils.TestAddress("payment-payer")

	testing.SetRealm(testing.NewUserRealm(addr))
	testing.SetOriginCaller(addr)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", 12))) // invalid

	uassert.AbortsContains(t, ErrInvalidPayment.Error(), func() {
		Register(cross, "nym-yankee345")
	})
}

// Audit finding #11: the OriginSend() payment check is only trustworthy
// when the direct caller is a pure EOA (IsUserCall). Any intermediate
// code realm can attach -send to the tx, keep the coins, and call
// Register; OriginSend() would still describe the envelope but namereg
// would receive nothing. The EOA-only guard must reject all such callers.
func TestRegister_IntermediateCodeRealmRejected(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/evil/wrapper"))
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))

	uassert.AbortsWithMessage(t, ErrNonUserCall.Error(), func() {
		Register(cross, "nym-whisky345")
	})
}

// Audit finding #13: the OriginSend mismatch error must report the
// CURRENT registerPrice, not the value frozen at package init time.
func TestRegister_PaymentErrorReflectsCurrentPrice(t *testing.T) {
	oldPrice := registerPrice
	defer func() { registerPrice = oldPrice }()

	registerPrice = 5_000_000

	addr := testutils.TestAddress("price-payer")
	testing.SetRealm(testing.NewUserRealm(addr))
	testing.SetOriginCaller(addr)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", 1_000_000))) // wrong

	uassert.AbortsContains(t, "5000000", func() {
		Register(cross, "nym-xrayfox678")
	})

	uassert.AbortsContains(t, ErrInvalidPayment.Error(), func() {
		Register(cross, "nym-xrayfox678")
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
		urequire.NotPanics(t, func() { ProposeNewRegisterPrice(MinRegisterPrice) })
	})

	t.Run("above floor is accepted", func(t *testing.T) {
		urequire.NotPanics(t, func() { ProposeNewRegisterPrice(1_000_000_000) })
	})
}
