package metadata

import (
	"testing"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

const (
	collectionName   = "MetaToken"
	collectionSymbol = "MTOK"
)

// newTestStack creates a core grc721.Token plus a metadata.PrivateLedger
// attached to it. The same-realm cross promotes the test's EOA-origin cur
// into a CodeRealm cur so grc721.NewToken's IsCurrent + non-empty PkgPath
// checks pass.
func newTestStack(_ int, rlm realm) (tok *grc721.Token, coreLedger *grc721.PrivateLedger, mLedger *PrivateLedger) {
	func(cur realm) {
		tok, coreLedger = grc721.NewToken(0, cur, collectionName, collectionSymbol)
	}(cross(rlm))
	mLedger = NewPrivateLedger(tok)
	return
}

func sampleMetadata() Metadata {
	return Metadata{
		Image:       "ipfs://example/1.png",
		Name:        "Sample #1",
		Description: "test fixture",
		Attributes: []Trait{
			{TraitType: "level", Value: "1"},
		},
	}
}

func TestNewPrivateLedger(cur realm, t *testing.T) {
	tok, coreLedger, mLedger := newTestStack(0, cur)
	uassert.True(t, tok != nil, "token should not be nil")
	uassert.True(t, coreLedger != nil, "core ledger should not be nil")
	uassert.True(t, mLedger != nil, "metadata ledger should not be nil")
}

func TestPrivateLedger_SetTokenMetadata_Admin(cur realm, t *testing.T) {
	_, coreLedger, mLedger := newTestStack(0, cur)
	owner := testutils.TestAddress("owner")

	uassert.NoError(t, coreLedger.Mint(owner, grc721.TokenID("1")))
	uassert.NoError(t, mLedger.SetTokenMetadata(grc721.TokenID("1"), sampleMetadata()))

	got, err := mLedger.TokenMetadata(grc721.TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, "ipfs://example/1.png", got.Image)
	uassert.Equal(t, "Sample #1", got.Name)
}

func TestPrivateLedger_SetTokenMetadata_MissingToken(cur realm, t *testing.T) {
	_, _, mLedger := newTestStack(0, cur)
	err := mLedger.SetTokenMetadata(grc721.TokenID("ghost"), sampleMetadata())
	uassert.ErrorIs(t, err, grc721.ErrInvalidTokenId)
}

func TestTokenMetadata_NotSet(cur realm, t *testing.T) {
	_, coreLedger, mLedger := newTestStack(0, cur)
	owner := testutils.TestAddress("owner")

	uassert.NoError(t, coreLedger.Mint(owner, grc721.TokenID("1")))

	_, err := mLedger.TokenMetadata(grc721.TokenID("1"))
	uassert.ErrorIs(t, err, ErrMetadataNotSet)
}

func TestTokenMetadata_MissingToken(cur realm, t *testing.T) {
	_, _, mLedger := newTestStack(0, cur)
	_, err := mLedger.TokenMetadata(grc721.TokenID("ghost"))
	uassert.ErrorIs(t, err, grc721.ErrInvalidTokenId)
}

func TestPrivateLedger_SetTokenURI_Admin(cur realm, t *testing.T) {
	_, coreLedger, mLedger := newTestStack(0, cur)
	owner := testutils.TestAddress("owner")
	uri := "ipfs://example/1"

	uassert.NoError(t, coreLedger.Mint(owner, grc721.TokenID("1")))
	uassert.NoError(t, mLedger.SetTokenURI(grc721.TokenID("1"), grc721.TokenURI(uri)))

	got, err := mLedger.TokenURI(grc721.TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, uri, got)
}

func TestPrivateLedger_SetTokenURI_MissingToken(cur realm, t *testing.T) {
	_, _, mLedger := newTestStack(0, cur)
	err := mLedger.SetTokenURI(grc721.TokenID("ghost"), grc721.TokenURI("ipfs://x"))
	uassert.ErrorIs(t, err, grc721.ErrInvalidTokenId)
}

func TestTokenURI_NotSet(cur realm, t *testing.T) {
	_, _, mLedger := newTestStack(0, cur)
	_, err := mLedger.TokenURI(grc721.TokenID("1"))
	uassert.ErrorIs(t, err, grc721.ErrInvalidTokenId)
}

func TestLedger_Burn_ClearsURIAndMetadata(cur realm, t *testing.T) {
	_, coreLedger, mLedger := newTestStack(0, cur)
	owner := testutils.TestAddress("owner")
	tid := grc721.TokenID("1")

	uassert.NoError(t, coreLedger.Mint(owner, tid))
	uassert.NoError(t, mLedger.SetTokenURI(tid, grc721.TokenURI("ipfs://example/1")))
	uassert.NoError(t, mLedger.SetTokenMetadata(tid, sampleMetadata()))

	// Host realm contract: extension Burn runs alongside core Burn.
	uassert.NoError(t, coreLedger.Burn(tid))
	mLedger.Burn(tid)

	_, err := mLedger.TokenURI(tid)
	uassert.ErrorIs(t, err, grc721.ErrInvalidTokenId)

	_, err = mLedger.TokenMetadata(tid)
	uassert.ErrorIs(t, err, grc721.ErrInvalidTokenId)
}

func TestNewPrivateLedger_NilToken(cur realm, t *testing.T) {
	defer func() {
		r := recover()
		uassert.True(t, r != nil, "NewPrivateLedger must panic on nil token")
	}()
	NewPrivateLedger(nil)
}

func TestLedger_Burn_Idempotent(cur realm, t *testing.T) {
	_, coreLedger, mLedger := newTestStack(0, cur)
	owner := testutils.TestAddress("owner")
	tid := grc721.TokenID("1")

	uassert.NoError(t, coreLedger.Mint(owner, tid))

	mLedger.Burn(tid)

	uassert.NoError(t, mLedger.SetTokenURI(tid, grc721.TokenURI("ipfs://x")))
	mLedger.Burn(tid)
	mLedger.Burn(tid)

	_, err := mLedger.TokenURI(tid)
	uassert.ErrorIs(t, err, grc721.ErrInvalidTokenId)
}

func TestImpersonateTeller_InvalidAddress(cur realm, t *testing.T) {
	_, _, mLedger := newTestStack(0, cur)

	defer func() {
		r := recover()
		uassert.True(t, r != nil, "ImpersonateTeller must panic on empty addr")
		uassert.True(t, r == grc721.ErrInvalidAddress,
			"ImpersonateTeller must panic with grc721.ErrInvalidAddress")
	}()
	mLedger.ImpersonateTeller(address(""))
}
