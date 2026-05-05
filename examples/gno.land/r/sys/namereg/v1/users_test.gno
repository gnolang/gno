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
		"nym-bravo123",         // 5-char stem (minimum)
		"nym-tango456",         // 5-char stem
		"nym-victor789",        // 6-char stem
		"nym-romeoecho012",     // 10-char stem
		"nym-mikefoxhote345",   // 11-char stem
		"nym-mikefoxhotenp678", // 13-char stem (maximum)
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
		"",                      // empty
		"    ",                  // whitespace
		"alice123",              // missing nym- prefix
		"usr-alice123",          // wrong prefix
		"nym-",                  // prefix only
		"nym-abc123",            // stem too short (3 chars, need ≥5)
		"nym-abcd123",           // stem too short (4 chars)
		"nym-abcdefghijklmn123", // stem too long (14 chars, need ≤13)
		"nym-Alice123",          // uppercase in stem
		"nym-al1ce123",          // digit in stem
		"nym-alice12",           // only 2 trailing digits
		"nym-alice1234",         // 4 trailing digits (extra digit in stem too long)
		"nym-alice&#($)123",     // special chars in stem
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
		"nym-gnoblah123",    // gno prefix
		"nym-gnomeland456",  // gno prefix
		"nym-glasgow012",    // gl prefix (most-confusable with bech32 g1...)
		"nym-atomic123",     // atom prefix
		"nym-atonex123",     // atone prefix
		"nym-photons456",    // photon prefix
		"nym-cosmoswide789", // cosmos prefix
	}

	for _, username := range cases {
		uassert.AbortsWithMessage(t, ErrReservedPrefix.Error(), func() {
			Register(cross, username)
		})
	}
}

// gi is intentionally NOT in reservedPrefixes (i is visually distinct
// enough from 1/l that legitimate gi* names should be registerable).
// Phishing protection for the gi/gl visual class is enforced by
// canonical-collision detection in r/sys/users — see
// TestRegister_CanonicalCollision.
func TestRegister_GiPrefix_Allowed(t *testing.T) {
	addr := testutils.TestAddress("gi-prefix-allowed")
	testing.SetRealm(testing.NewUserRealm(addr))
	testing.SetOriginCaller(addr)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))

	urequire.NotPanics(t, func() {
		Register(cross, "nym-gillette789")
	})
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
	// First register a name. Same-digits confusable variant must collide
	// in the unified susers canonical store.
	addr1 := testutils.TestAddress("collision-1")
	testing.SetRealm(testing.NewUserRealm(addr1))
	testing.SetOriginCaller(addr1)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
	urequire.NotPanics(t, func() {
		Register(cross, "nym-balloon123")
	})

	// Then attempt to register a canonical-equivalent variant — same
	// canonical full name (after l→i, etc.). Must reject from susers.
	addr2 := testutils.TestAddress("collision-2")
	testing.SetRealm(testing.NewUserRealm(addr2))
	testing.SetOriginCaller(addr2)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
	uassert.AbortsWithMessage(t, susers.ErrCanonicalCollision.Error(), func() {
		Register(cross, "nym-baiioon123")
	})

	// Reverse direction also blocked.
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
	uassert.AbortsWithMessage(t, susers.ErrCanonicalCollision.Error(), func() {
		Register(cross, "nym-papallo789")
	})
}

// Decision #2 in Option B: store key is full canonical username (not
// stem). Same alpha stem with different digit suffixes coexists — does
// not collide.
func TestRegister_SameStemDifferentDigits_Coexist(t *testing.T) {
	addrA := testutils.TestAddress("stem-coexist-A")
	testing.SetRealm(testing.NewUserRealm(addrA))
	testing.SetOriginCaller(addrA)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
	urequire.NotPanics(t, func() {
		Register(cross, "nym-foolbar000")
	})

	addrB := testutils.TestAddress("stem-coexist-B")
	testing.SetRealm(testing.NewUserRealm(addrB))
	testing.SetOriginCaller(addrB)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
	urequire.NotPanics(t, func() {
		Register(cross, "nym-foolbar999")
	})
}

func TestRegister_TakenUsername(t *testing.T) {
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", registerPrice)))
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("taken-payer")))

	username := "nym-zulutwo567"

	urequire.NotPanics(t, func() {
		Register(cross, username)
	})

	// Re-registration of the same exact name is rejected by susers via
	// ErrNameTaken (the nameStore.Has check precedes the canonical check).
	uassert.AbortsWithMessage(t, susers.ErrNameTaken.Error(), func() {
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
// negative prices. A passed proposal to set a negative price would
// have been arithmetically nonsense; reject below MinRegisterPrice
// (currently 0) at proposal-creation time so governance can never
// underflow the floor.
func TestProposeNewRegisterPrice_floor(t *testing.T) {
	t.Run("zero is accepted (free registration is the default)", func(t *testing.T) {
		urequire.NotPanics(t, func() { ProposeNewRegisterPrice(0) })
	})

	t.Run("negative is rejected", func(t *testing.T) {
		urequire.PanicsWithMessage(t,
			"price below floor: -1 ugnot < 0 ugnot (MinRegisterPrice)",
			func() { ProposeNewRegisterPrice(-1) })
	})

	t.Run("at floor is accepted", func(t *testing.T) {
		urequire.NotPanics(t, func() { ProposeNewRegisterPrice(MinRegisterPrice) })
	})

	t.Run("above floor is accepted", func(t *testing.T) {
		urequire.NotPanics(t, func() { ProposeNewRegisterPrice(1_000_000_000) })
	})
}
