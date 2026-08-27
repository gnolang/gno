package collection

import (
	"testing"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/demo/tokens/grc721/metadata"
	"gno.land/p/demo/tokens/grc721/royalty"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// Same-realm cross: NewToken requires rlm.IsCurrent().
func newNFT(name, symbol string, id seqid.ID, rlm realm) (nft *NFT) {
	func(cur realm) {
		nft, _ = New(name, symbol, id, cur)
	}(cross(rlm))

	return
}

func newTokenLedger(name, symbol string, id seqid.ID, rlm realm) (tok *grc721.Token, led *grc721.PrivateLedger) {
	func(cur realm) {
		tok, led = grc721.NewToken(name, symbol, id, cur)
	}(cross(rlm))

	return
}

func TestNewConstructors(cur realm, t *testing.T) {
	t.Run("New builds an NFT with a core token", func(t *testing.T) {
		nft := newNFT("Foo NFT", "FNFT", 0, cur)
		uassert.Equal(t, "Foo NFT", nft.Token().GetName())
		uassert.Equal(t, "FNFT", nft.Token().GetSymbol())
		urequire.True(t, nft.Collection() != nil, "collection present")
		urequire.True(t, nft.ledger != nil, "ledger retained")
		uassert.Equal(t, nft.Token().ID(), nft.Collection().Token().ID())
	})

	t.Run("CallerTeller returns the canonical teller", func(t *testing.T) {
		nft := newNFT("Foo NFT", "FNFT", 0, cur)
		teller := nft.CallerTeller()
		urequire.True(t, teller != nil, "teller present")
		uassert.True(t, grc721.IsCanonicalTeller(teller))
	})

	t.Run("Wrap builds an NFT around a token/ledger pair", func(t *testing.T) {
		tok, led := newTokenLedger("Bar", "BAR", 1, cur)
		nft := Wrap(tok, led)
		uassert.Equal(t, "Bar", nft.Token().GetName())
		uassert.Equal(t, tok.ID(), nft.Token().ID())
		uassert.Equal(t, tok.ID(), nft.Collection().Token().ID())
	})

	t.Run("Wrap rejects a mismatched token/ledger pair", func(t *testing.T) {
		tokA, _ := newTokenLedger("A", "AAA", 3, cur)
		_, ledB := newTokenLedger("B", "BBB", 4, cur)
		uassert.PanicsWithMessage(t, cur, "collection: ledger does not belong to this token", func() {
			Wrap(tokA, ledB)
		})
	})

	panics := []struct {
		name string
		run  func()
	}{
		{"Wrap nil token panics", func() { Wrap(nil, nil) }},
		{"Wrap nil ledger panics", func() {
			tok, _ := newTokenLedger("Baz", "BAZ", 2, cur)
			Wrap(tok, nil)
		}},
	}
	for _, tt := range panics {
		t.Run(tt.name, func(t *testing.T) {
			uassert.PanicsWithMessage(t, cur, "collection: nil token or ledger", tt.run)
		})
	}
}

func TestNewCollection(cur realm, t *testing.T) {
	t.Run("NewCollection nil ledger panics", func(t *testing.T) {
		uassert.PanicsWithMessage(t, cur, "collection: nil ledger", func() {
			NewCollection(nil)
		})
	})

	t.Run("NewCollection rejects a ledger with no token", func(t *testing.T) {
		uassert.PanicsWithMessage(t, cur, "collection: ledger has no token", func() {
			NewCollection(&grc721.PrivateLedger{})
		})
	})

	t.Run("NewCollection with no extensions builds an empty collection", func(t *testing.T) {
		tok, led := newTokenLedger("Foo", "FOO", 0, cur)
		c := NewCollection(led)
		uassert.Equal(t, tok.ID(), c.Token().ID())
		uassert.Equal(t, 0, len(c.Kinds()))
	})

	t.Run("NewCollection takes every registered view", func(t *testing.T) {
		_, led := newTokenLedger("Foo", "FOO", 0, cur)
		metadata.NewMetadata(led)
		royalty.NewRoyalty(led, 1000)
		c := NewCollection(led)
		uassert.Equal(t, 2, len(c.Kinds()))
		uassert.True(t, c.HasExtension(metadata.Kind))
		uassert.True(t, c.HasExtension(royalty.Kind))
	})

	t.Run("NewCollection skips a kind registered without a view", func(t *testing.T) {
		_, led := newTokenLedger("Foo", "FOO", 0, cur)
		led.RegisterExtension(&viewlessExtension{}, nil)
		c := NewCollection(led)
		uassert.Equal(t, 0, len(c.Kinds()))
	})
}

// viewlessExtension has hooks but publishes nothing.
type viewlessExtension struct{}

func (v *viewlessExtension) ExtensionKind() string { return "viewless" }

func (v *viewlessExtension) OnMint(to address, tid grc721.TokenID) {}

func (v *viewlessExtension) OnTransfer(from, to address, tid grc721.TokenID) {}

func (v *viewlessExtension) OnBurn(tid grc721.TokenID) {}

func TestExtensionAccessors(cur realm, t *testing.T) {
	nft := newNFT("Foo NFT", "FNFT", 0, cur)
	metadata.NewMetadata(nft.ledger)
	royalty.NewRoyalty(nft.ledger, 1000)

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"Kinds returns sorted registered kinds", func(t *testing.T) {
			ks := nft.Collection().Kinds()
			urequire.Equal(t, 2, len(ks))
			uassert.Equal(t, metadata.Kind, ks[0])
			uassert.Equal(t, royalty.Kind, ks[1])
		}},
		{"HasExtension true for registered kind", func(t *testing.T) {
			uassert.True(t, nft.Collection().HasExtension(metadata.Kind))
		}},
		{"HasExtension false for absent kind", func(t *testing.T) {
			uassert.False(t, nft.Collection().HasExtension("enumerable"))
		}},
		{"NFT.Extension returns the registered typed view", func(t *testing.T) {
			v, ok := nft.Extension(metadata.Kind)
			urequire.True(t, ok, "metadata found")
			uassert.Equal(t, metadata.Kind, v.(*metadata.Metadata).ExtensionKind())
		}},
		{"Collection.Extension returns the registered typed view", func(t *testing.T) {
			v, ok := nft.Collection().Extension(royalty.Kind)
			urequire.True(t, ok, "royalty found")
			uassert.Equal(t, royalty.Kind, v.(*royalty.Royalty).ExtensionKind())
		}},
		{"Extension absent kind returns (nil,false)", func(t *testing.T) {
			v, ok := nft.Extension("enumerable")
			uassert.False(t, ok)
			urequire.True(t, v == nil, "nil view")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestExtensionViewsComeFromTheLedger(cur realm, t *testing.T) {
	nft := newNFT("Foo", "FOO", 0, cur)
	roy, _ := royalty.NewRoyalty(nft.ledger, 1000)

	v, ok := nft.Collection().Extension(royalty.Kind)
	urequire.True(t, ok, "royalty found")
	uassert.True(t, v.(*royalty.Royalty) == roy, "the ledger's own view, not a caller's")

	// a view built on another ledger cannot reach this collection: there is no
	// parameter to pass it through, and this token's ledger never saw it
	other := newNFT("Other", "OTH", 1, cur)
	royOther, _ := royalty.NewRoyalty(other.ledger, 1000)
	v, _ = nft.Collection().Extension(royalty.Kind)
	uassert.False(t, v.(*royalty.Royalty) == royOther)
}

func TestKindsOrderIsIndependentOfRegistrationOrder(cur realm, t *testing.T) {
	nft := newNFT("Foo NFT", "FNFT", 0, cur)
	royalty.NewRoyalty(nft.ledger, 1000) // registered first, sorts second
	metadata.NewMetadata(nft.ledger)

	kinds := nft.Token().ExtensionKinds()
	urequire.Equal(t, 2, len(kinds))
	uassert.Equal(t, royalty.Kind, kinds[0], "the token keeps registration order")

	ks := nft.Collection().Kinds()
	urequire.Equal(t, 2, len(ks))
	uassert.Equal(t, metadata.Kind, ks[0], "Collection.Kinds is sorted")
	uassert.Equal(t, royalty.Kind, ks[1])
}

func TestPublishedCollectionsTrackTheirLedger(cur realm, t *testing.T) {
	tok, led := newTokenLedger("Foo", "FOO", 0, cur)
	standalone := NewCollection(led)
	published := Wrap(tok, led).Collection() // what a registry stores

	metadata.NewMetadata(led) // a later transaction

	rewrapped := Wrap(tok, led).Collection() // the issuer re-wraps its persisted pair
	uassert.False(t, published == rewrapped, "re-wrapping builds a new handle")
	uassert.Equal(t, 1, len(standalone.Kinds()))
	uassert.Equal(t, 1, len(published.Kinds()))
	uassert.Equal(t, 1, len(rewrapped.Kinds()))
}

func TestLateRegistrationReachesAPublishedCollection(cur realm, t *testing.T) {
	nft := newNFT("Foo", "FOO", 0, cur)
	published := nft.Collection() // the pointer a registry would already hold
	urequire.Equal(t, 0, len(published.Kinds()))

	metadata.NewMetadata(nft.ledger)

	uassert.True(t, published == nft.Collection())
	urequire.Equal(t, 1, len(published.Kinds()))
	uassert.True(t, published.HasExtension(metadata.Kind))
}

func TestMintBurn(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	artist := testutils.TestAddress("artist")

	nft := newNFT("Foo NFT", "FNFT", 0, cur)
	meta, metaLedger := metadata.NewMetadata(nft.ledger)
	_, royLedger := royalty.NewRoyalty(nft.ledger, 1000)

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"Mint succeeds and bumps supply", func(t *testing.T) {
			urequire.NoError(t, nft.Mint(alice, "1"))
			uassert.Equal(t, int64(1), nft.Token().TotalSupply())
		}},
		{"Mint invalid address returns ErrInvalidAddress", func(t *testing.T) {
			err := nft.Mint(address(""), "2")
			uassert.ErrorIs(t, err, grc721.ErrInvalidAddress)
		}},
		{"Mint duplicate id returns ErrTokenIdAlreadyExists", func(t *testing.T) {
			err := nft.Mint(alice, "1")
			uassert.ErrorIs(t, err, grc721.ErrTokenIdAlreadyExists)
		}},
		{"per-token extension state set through ledgers", func(t *testing.T) {
			urequire.NoError(t, metaLedger.SetTokenURI("1", "ipfs://uri-1"))
			urequire.NoError(t, royLedger.SetDefaultRoyalty(artist, 500))
			uri, err := meta.TokenURI("1")
			urequire.NoError(t, err)
			uassert.Equal(t, "ipfs://uri-1", uri)
		}},
		{"Burn nonexistent id returns ErrInvalidTokenId", func(t *testing.T) {
			err := nft.Burn("999")
			uassert.ErrorIs(t, err, grc721.ErrInvalidTokenId)
		}},
		{"Burn succeeds and fans out to extension hooks", func(t *testing.T) {
			urequire.NoError(t, nft.Burn("1"))
			uassert.Equal(t, int64(0), nft.Token().TotalSupply())
			uassert.False(t, meta.HasTokenURI("1"))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestNewInvalidAborts(cur realm, t *testing.T) {
	tests := []struct {
		name   string
		symbol string
		msg    string
	}{
		{"invalid symbol aborts", "bad symbol!", "invalid token symbol"},
		{"empty symbol aborts", "", "invalid token symbol"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urequire.AbortsContains(t, cur, tt.msg, func() {
				newNFT("Foo NFT", tt.symbol, 0, cur)
			})
		})
	}
}
