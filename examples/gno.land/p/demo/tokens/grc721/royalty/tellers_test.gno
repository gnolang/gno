package royalty

import (
	"testing"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

// evilWrap embeds Teller; IsCanonicalTeller must reject it.
type evilWrap struct {
	Teller
}

func TestIsCanonicalTeller(cur realm, t *testing.T) {
	_, _, rLedger := newTestStack(0, cur, 1000)
	legit := rLedger.CallerTeller()

	uassert.True(t, IsCanonicalTeller(legit),
		"canonical *fnTeller from CallerTeller must pass")

	evil := &evilWrap{Teller: legit}
	uassert.False(t, IsCanonicalTeller(evil),
		"foreign type embedding a canonical Teller must be rejected")
}

func TestCallerTeller_RoyaltyInfo(cur realm, t *testing.T) {
	_, ledger, rLedger := newTestStack(0, cur, 1000)
	owner := testutils.TestAddress("owner")
	payee := testutils.TestAddress("payee")

	uassert.NoError(t, ledger.Mint(owner, grc721.TokenID("1")))
	uassert.NoError(t, rLedger.SetTokenRoyalty(grc721.TokenID("1"), RoyaltyInfo{
		PaymentAddress: payee,
		Bps:            300,
	}))

	teller := rLedger.CallerTeller()
	addr, amount, err := teller.RoyaltyInfo(grc721.TokenID("1"), 10000)
	uassert.NoError(t, err)
	uassert.Equal(t, payee.String(), addr.String())
	uassert.Equal(t, int64(300), amount)
}

func TestRealmTeller_Smoke(cur realm, t *testing.T) {
	_, _, rLedger := newTestStack(0, cur, 1000)

	var teller Teller
	func(cur realm) {
		teller = rLedger.RealmTeller(0, cur)
	}(cross(cur))
	uassert.True(t, teller != nil, "teller should not be nil")
}

// TestCoexistence_SameToken demonstrates that metadata and royalty can attach
// to the same core token simultaneously — the whole point of the attach-style
// API. The assertion is structural: both ledgers operate on the same tid.
func TestCoexistence_SameToken(cur realm, t *testing.T) {
	tok, coreLedger, rLedger := newTestStack(0, cur, 1000)
	owner := testutils.TestAddress("owner")
	payee := testutils.TestAddress("payee")

	uassert.NoError(t, coreLedger.Mint(owner, grc721.TokenID("1")))

	// Royalty on the shared token.
	uassert.NoError(t, rLedger.SetTokenRoyalty(grc721.TokenID("1"), RoyaltyInfo{
		PaymentAddress: payee,
		Bps:            500,
	}))

	// Core ownership and royalty info both resolve via the same *grc721.Token.
	gotOwner, err := tok.OwnerOf(grc721.TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, owner.String(), gotOwner.String())

	addr, amount, err := rLedger.RoyaltyInfo(grc721.TokenID("1"), 10000)
	uassert.NoError(t, err)
	uassert.Equal(t, payee.String(), addr.String())
	uassert.Equal(t, int64(500), amount)
}

func TestReadonlyTeller_RoyaltyInfo(cur realm, t *testing.T) {
	_, ledger, rLedger := newTestStack(0, cur, 1000)
	owner := testutils.TestAddress("owner")
	payee := testutils.TestAddress("payee")

	uassert.NoError(t, ledger.Mint(owner, grc721.TokenID("1")))
	uassert.NoError(t, rLedger.SetTokenRoyalty(grc721.TokenID("1"), RoyaltyInfo{
		PaymentAddress: payee,
		Bps:            500,
	}))

	teller := rLedger.ReadonlyTeller()
	addr, amount, err := teller.RoyaltyInfo(grc721.TokenID("1"), 10000)
	uassert.NoError(t, err)
	uassert.Equal(t, payee.String(), addr.String())
	uassert.Equal(t, int64(500), amount)
}
