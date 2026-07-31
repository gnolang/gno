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
		nft = New(name, symbol, id, cur)
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
		urequire.True(t, nft.Ledger() != nil, "ledger present")
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
	t.Run("NewCollection nil token panics", func(t *testing.T) {
		uassert.PanicsWithMessage(t, cur, "collection: nil token", func() {
			NewCollection(nil)
		})
	})

	t.Run("NewCollection with no views builds an empty collection", func(t *testing.T) {
		tok, _ := newTokenLedger("Foo", "FOO", 0, cur)
		c := NewCollection(tok)
		uassert.Equal(t, tok.ID(), c.Token().ID())
		uassert.Equal(t, 0, len(c.Kinds()))
	})

	t.Run("NewCollection skips a nil view", func(t *testing.T) {
		tok, _ := newTokenLedger("Foo", "FOO", 0, cur)
		var nilView grc721.ExtensionView
		c := NewCollection(tok, nilView)
		uassert.Equal(t, 0, len(c.Kinds()))
	})

	t.Run("NewCollection attaches matching views", func(t *testing.T) {
		tok, led := newTokenLedger("Foo", "FOO", 0, cur)
		meta, _ := metadata.NewMetadata(led)
		roy, _ := royalty.NewRoyalty(led, 1000)
		c := NewCollection(tok, meta, roy)
		uassert.Equal(t, 2, len(c.Kinds()))
		uassert.True(t, c.HasExtension(metadata.Kind))
		uassert.True(t, c.HasExtension(royalty.Kind))
	})

	t.Run("NewCollection rejects a view from a different token", func(t *testing.T) {
		tok, _ := newTokenLedger("Foo", "FOO", 0, cur)
		_, otherLed := newTokenLedger("Other", "OTH", 1, cur)
		royOther, _ := royalty.NewRoyalty(otherLed, 1000)
		uassert.PanicsWithMessage(t, cur, "collection: view does not belong to this token", func() {
			NewCollection(tok, royOther)
		})
	})

	t.Run("NewCollection rejects a duplicate kind", func(t *testing.T) {
		tok, led := newTokenLedger("Foo", "FOO", 0, cur)
		roy, _ := royalty.NewRoyalty(led, 1000)
		uassert.PanicsWithMessage(t, cur, "collection: duplicate extension kind: royalty", func() {
			NewCollection(tok, roy, roy)
		})
	})
}

func TestExtensionAccessors(cur realm, t *testing.T) {
	nft := newNFT("Foo NFT", "FNFT", 0, cur)
	meta, _ := metadata.NewMetadata(nft.Ledger())
	roy, _ := royalty.NewRoyalty(nft.Ledger(), 1000)
	nft.Attach(meta).Attach(roy)

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"Kinds returns sorted attached kinds", func(t *testing.T) {
			ks := nft.Collection().Kinds()
			urequire.Equal(t, 2, len(ks))
			uassert.Equal(t, metadata.Kind, ks[0])
			uassert.Equal(t, royalty.Kind, ks[1])
		}},
		{"HasExtension true for attached kind", func(t *testing.T) {
			uassert.True(t, nft.Collection().HasExtension(metadata.Kind))
		}},
		{"HasExtension false for absent kind", func(t *testing.T) {
			uassert.False(t, nft.Collection().HasExtension("enumerable"))
		}},
		{"NFT.Extension returns attached typed view", func(t *testing.T) {
			v, ok := nft.Extension(metadata.Kind)
			urequire.True(t, ok, "metadata found")
			uassert.Equal(t, metadata.Kind, v.(*metadata.Metadata).ExtensionKind())
		}},
		{"Collection.Extension returns attached typed view", func(t *testing.T) {
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

func TestAttach(cur realm, t *testing.T) {
	t.Run("Attach chains and records the view", func(t *testing.T) {
		nft := newNFT("Foo", "FOO", 0, cur)
		roy, _ := royalty.NewRoyalty(nft.Ledger(), 1000)
		got := nft.Attach(roy)
		urequire.True(t, got == nft, "Attach returns the NFT for chaining")
		uassert.True(t, nft.Collection().HasExtension(royalty.Kind))
	})

	t.Run("Attach skips a nil view", func(t *testing.T) {
		nft := newNFT("Foo", "FOO", 0, cur)
		var nilView grc721.ExtensionView
		nft.Attach(nilView)
		uassert.Equal(t, 0, len(nft.Collection().Kinds()))
	})

	t.Run("Attach rejects a view from a different token", func(t *testing.T) {
		nft := newNFT("Foo", "FOO", 0, cur)
		other := newNFT("Other", "OTH", 1, cur)
		royOther, _ := royalty.NewRoyalty(other.Ledger(), 1000)
		uassert.PanicsWithMessage(t, cur, "collection: view does not belong to this token", func() {
			nft.Attach(royOther)
		})
	})

	t.Run("Attach rejects a duplicate kind", func(t *testing.T) {
		nft := newNFT("Foo", "FOO", 0, cur)
		roy, _ := royalty.NewRoyalty(nft.Ledger(), 1000)
		nft.Attach(roy)
		uassert.PanicsWithMessage(t, cur, "collection: duplicate extension kind: royalty", func() {
			nft.Attach(roy)
		})
	})
}

func TestMintBurn(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	artist := testutils.TestAddress("artist")

	nft := newNFT("Foo NFT", "FNFT", 0, cur)
	meta, metaLedger := metadata.NewMetadata(nft.Ledger())
	roy, royLedger := royalty.NewRoyalty(nft.Ledger(), 1000)
	nft.Attach(meta).Attach(roy)

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
