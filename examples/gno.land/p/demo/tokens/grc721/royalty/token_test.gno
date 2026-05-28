package royalty

import (
	"testing"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

const (
	collectionName   = "RoyaltyToken"
	collectionSymbol = "RTOK"
)

// newTestStack creates a core grc721.Token plus a royalty.PrivateLedger
// attached to it. The same-realm cross promotes the test's EOA-origin cur
// into a CodeRealm cur so grc721.NewToken's IsCurrent + non-empty PkgPath
// checks pass.
func newTestStack(_ int, rlm realm, maxBps int64) (tok *grc721.Token, coreLedger *grc721.PrivateLedger, rLedger *PrivateLedger) {
	func(cur realm) {
		tok, coreLedger = grc721.NewToken(0, cur, collectionName, collectionSymbol)
	}(cross(rlm))
	rLedger = NewPrivateLedger(tok, maxBps)
	return
}

func TestNewPrivateLedger(cur realm, t *testing.T) {
	tok, coreLedger, rLedger := newTestStack(0, cur, 1000)
	uassert.True(t, tok != nil, "token should not be nil")
	uassert.True(t, coreLedger != nil, "core ledger should not be nil")
	uassert.True(t, rLedger != nil, "royalty ledger should not be nil")
	uassert.Equal(t, int64(1000), rLedger.MaxBps())
}

func TestNewPrivateLedger_DefaultMaxBps(cur realm, t *testing.T) {
	_, _, rLedger := newTestStack(0, cur, 0)
	uassert.Equal(t, int64(MaxBps), rLedger.MaxBps())
}

func TestPrivateLedger_SetTokenRoyalty_Admin(cur realm, t *testing.T) {
	_, ledger, rLedger := newTestStack(0, cur, 1000)
	owner := testutils.TestAddress("owner")
	payee := testutils.TestAddress("payee")

	uassert.NoError(t, ledger.Mint(owner, grc721.TokenID("1")))
	uassert.NoError(t, rLedger.SetTokenRoyalty(grc721.TokenID("1"), RoyaltyInfo{
		PaymentAddress: payee,
		Bps:            500, // 5%
	}))

	addr, amount, err := rLedger.RoyaltyInfo(grc721.TokenID("1"), 10000)
	uassert.NoError(t, err)
	uassert.Equal(t, payee.String(), addr.String())
	uassert.Equal(t, int64(500), amount) // 10000 * 500 / 10000 = 500
}

func TestPrivateLedger_SetTokenRoyalty_BpsOutOfRange(cur realm, t *testing.T) {
	_, ledger, rLedger := newTestStack(0, cur, 1000)
	owner := testutils.TestAddress("owner")
	payee := testutils.TestAddress("payee")

	uassert.NoError(t, ledger.Mint(owner, grc721.TokenID("1")))

	err := rLedger.SetTokenRoyalty(grc721.TokenID("1"), RoyaltyInfo{
		PaymentAddress: payee,
		Bps:            5000, // exceeds maxBps=1000
	})
	uassert.ErrorIs(t, err, ErrInvalidRoyaltyBps)
}

func TestPrivateLedger_SetTokenRoyalty_InvalidPayee(cur realm, t *testing.T) {
	_, ledger, rLedger := newTestStack(0, cur, 1000)
	owner := testutils.TestAddress("owner")

	uassert.NoError(t, ledger.Mint(owner, grc721.TokenID("1")))

	err := rLedger.SetTokenRoyalty(grc721.TokenID("1"), RoyaltyInfo{
		PaymentAddress: address(""),
		Bps:            500,
	})
	uassert.ErrorIs(t, err, ErrInvalidRoyaltyPaymentAddress)
}

func TestPrivateLedger_SetTokenRoyalty_MissingToken(cur realm, t *testing.T) {
	_, _, rLedger := newTestStack(0, cur, 1000)
	payee := testutils.TestAddress("payee")

	err := rLedger.SetTokenRoyalty(grc721.TokenID("ghost"), RoyaltyInfo{
		PaymentAddress: payee,
		Bps:            500,
	})
	uassert.ErrorIs(t, err, grc721.ErrInvalidTokenId)
}

func TestLedger_Burn_ClearsRoyalty(cur realm, t *testing.T) {
	_, ledger, rLedger := newTestStack(0, cur, 1000)
	owner := testutils.TestAddress("owner")
	payee := testutils.TestAddress("payee")
	tid := grc721.TokenID("1")

	uassert.NoError(t, ledger.Mint(owner, tid))
	uassert.NoError(t, rLedger.SetTokenRoyalty(tid, RoyaltyInfo{
		PaymentAddress: payee,
		Bps:            500,
	}))

	// Host realm contract: extension Burn runs alongside core Burn.
	uassert.NoError(t, ledger.Burn(tid))
	rLedger.Burn(tid)

	_, _, err := rLedger.RoyaltyInfo(tid, 10000)
	uassert.ErrorIs(t, err, grc721.ErrInvalidTokenId)
}

func TestRoyaltyInfo_FractionalPercent(cur realm, t *testing.T) {
	// EIP-2981 bps lets us represent fractional rates a percentage-based
	// ledger could not. Verify 2.5% (250 bps).
	_, ledger, rLedger := newTestStack(0, cur, 1000)
	owner := testutils.TestAddress("owner")
	payee := testutils.TestAddress("payee")

	uassert.NoError(t, ledger.Mint(owner, grc721.TokenID("1")))
	uassert.NoError(t, rLedger.SetTokenRoyalty(grc721.TokenID("1"), RoyaltyInfo{
		PaymentAddress: payee,
		Bps:            250,
	}))

	_, amount, err := rLedger.RoyaltyInfo(grc721.TokenID("1"), 100_000)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(2500), amount) // 100000 * 250 / 10000 = 2500
}

func TestNewPrivateLedger_NilToken(cur realm, t *testing.T) {
	defer func() {
		r := recover()
		uassert.True(t, r != nil, "NewPrivateLedger must panic on nil token")
	}()
	NewPrivateLedger(nil, 1000)
}

func TestNewPrivateLedger_NegativeMaxBps(cur realm, t *testing.T) {
	tok, _ := func() (*grc721.Token, *grc721.PrivateLedger) {
		var tok *grc721.Token
		var led *grc721.PrivateLedger
		func(cur realm) {
			tok, led = grc721.NewToken(0, cur, collectionName, collectionSymbol)
		}(cross(cur))
		return tok, led
	}()

	defer func() {
		r := recover()
		uassert.True(t, r != nil, "NewPrivateLedger must panic on negative maxBps")
		uassert.True(t, r == ErrInvalidMaxBps, "must panic with ErrInvalidMaxBps")
	}()
	NewPrivateLedger(tok, -1)
}

func TestNewPrivateLedger_MaxBpsOverflow(cur realm, t *testing.T) {
	tok, _ := func() (*grc721.Token, *grc721.PrivateLedger) {
		var tok *grc721.Token
		var led *grc721.PrivateLedger
		func(cur realm) {
			tok, led = grc721.NewToken(0, cur, collectionName, collectionSymbol)
		}(cross(cur))
		return tok, led
	}()

	defer func() {
		r := recover()
		uassert.True(t, r != nil, "NewPrivateLedger must panic when maxBps > MaxBps")
		uassert.True(t, r == ErrInvalidMaxBps, "must panic with ErrInvalidMaxBps")
	}()
	NewPrivateLedger(tok, MaxBps+1)
}

func TestRoyaltyInfo_MissingToken(cur realm, t *testing.T) {
	_, _, rLedger := newTestStack(0, cur, 1000)

	_, _, err := rLedger.RoyaltyInfo(grc721.TokenID("ghost"), 10000)
	uassert.ErrorIs(t, err, grc721.ErrInvalidTokenId)
}

func TestLedger_Burn_Idempotent(cur realm, t *testing.T) {
	_, ledger, rLedger := newTestStack(0, cur, 1000)
	owner := testutils.TestAddress("owner")
	payee := testutils.TestAddress("payee")
	tid := grc721.TokenID("1")

	uassert.NoError(t, ledger.Mint(owner, tid))

	rLedger.Burn(tid)

	uassert.NoError(t, rLedger.SetTokenRoyalty(tid, RoyaltyInfo{
		PaymentAddress: payee,
		Bps:            500,
	}))
	rLedger.Burn(tid)
	rLedger.Burn(tid)

	_, _, err := rLedger.RoyaltyInfo(tid, 10000)
	uassert.ErrorIs(t, err, grc721.ErrInvalidTokenId)
}

func TestImpersonateTeller_InvalidAddress(cur realm, t *testing.T) {
	_, _, rLedger := newTestStack(0, cur, 1000)

	defer func() {
		r := recover()
		uassert.True(t, r != nil, "ImpersonateTeller must panic on empty addr")
		uassert.True(t, r == grc721.ErrInvalidAddress,
			"ImpersonateTeller must panic with grc721.ErrInvalidAddress")
	}()
	rLedger.ImpersonateTeller(address(""))
}
