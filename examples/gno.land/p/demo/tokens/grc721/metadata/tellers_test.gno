package metadata

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
	_, _, mLedger := newTestStack(0, cur)
	legit := mLedger.CallerTeller()

	uassert.True(t, IsCanonicalTeller(legit),
		"canonical *fnTeller from CallerTeller must pass")

	evil := &evilWrap{Teller: legit}
	uassert.False(t, IsCanonicalTeller(evil),
		"foreign type embedding a canonical Teller must be rejected")
}

func TestCallerTeller_TokenURI(cur realm, t *testing.T) {
	_, coreLedger, mLedger := newTestStack(0, cur)
	alice := testutils.TestAddress("alice")
	uri := "ipfs://example/1"

	uassert.NoError(t, coreLedger.Mint(alice, grc721.TokenID("1")))
	uassert.NoError(t, mLedger.SetTokenURI(grc721.TokenID("1"), grc721.TokenURI(uri)))

	teller := mLedger.CallerTeller()
	got, err := teller.TokenURI(grc721.TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, uri, got)
}

func TestCallerTeller_TokenMetadata(cur realm, t *testing.T) {
	_, coreLedger, mLedger := newTestStack(0, cur)
	alice := testutils.TestAddress("alice")

	uassert.NoError(t, coreLedger.Mint(alice, grc721.TokenID("1")))
	uassert.NoError(t, mLedger.SetTokenMetadata(grc721.TokenID("1"), sampleMetadata()))

	teller := mLedger.CallerTeller()
	got, err := teller.TokenMetadata(grc721.TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, "Sample #1", got.Name)
}

func TestRealmTeller_Smoke(cur realm, t *testing.T) {
	_, _, mLedger := newTestStack(0, cur)

	var teller Teller
	func(cur realm) {
		teller = mLedger.RealmTeller(0, cur)
	}(cross(cur))
	uassert.True(t, teller != nil, "teller should not be nil")
}

func TestReadonlyTeller_Reads(cur realm, t *testing.T) {
	_, coreLedger, mLedger := newTestStack(0, cur)
	alice := testutils.TestAddress("alice")
	uri := "ipfs://example/1"

	uassert.NoError(t, coreLedger.Mint(alice, grc721.TokenID("1")))
	uassert.NoError(t, mLedger.SetTokenURI(grc721.TokenID("1"), grc721.TokenURI(uri)))
	uassert.NoError(t, mLedger.SetTokenMetadata(grc721.TokenID("1"), sampleMetadata()))

	teller := mLedger.ReadonlyTeller()

	gotURI, err := teller.TokenURI(grc721.TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, uri, gotURI)

	gotMeta, err := teller.TokenMetadata(grc721.TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, "Sample #1", gotMeta.Name)
}
