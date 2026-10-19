package metadata

import (
	"testing"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// newToken builds via same-realm cross (NewToken requires rlm.IsCurrent()).
func newToken(name, symbol string, id seqid.ID, rlm realm) (tok *grc721.Token, led *grc721.PrivateLedger) {
	func(cur realm) {
		tok, led = grc721.NewToken(name, symbol, id, cur)
	}(cross(rlm))
	return
}

func newMeta(id seqid.ID, rlm realm) (*grc721.PrivateLedger, *Metadata, *Ledger) {
	_, tokenLedger := newToken("Foo", "FOO", id, rlm)
	meta, metaLedger := NewMetadata(tokenLedger)
	return tokenLedger, meta, metaLedger
}

func TestNewMetadata(cur realm, t *testing.T) {
	tests := []struct {
		name      string
		nilLedger bool
	}{
		{"nil token ledger panics", true},
		{"valid token ledger constructs read view and ledger", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.nilLedger {
				uassert.PanicsWithMessage(t, cur, "metadata: nil core ledger", func() {
					NewMetadata(nil)
				})
				return
			}

			_, tokenLedger := newToken("Foo", "FOO", 0, cur)
			meta, metaLedger := NewMetadata(tokenLedger)

			uassert.True(t, meta != nil)
			uassert.True(t, metaLedger != nil)
			uassert.Equal(t, Kind, meta.ExtensionKind())
			uassert.Equal(t, "metadata", meta.ExtensionKind())
			uassert.Equal(t, tokenLedger.ReadToken().ID(), meta.TokenID())
		})
	}
}

func TestSetTokenURI(cur realm, t *testing.T) {
	tests := []struct {
		name    string
		mint    bool
		tid     grc721.TokenID
		uri     string
		wantErr error
	}{
		{"unminted token returns ErrTokenNotMinted", false, "404", "ipfs://x", ErrTokenNotMinted},
		{"minted token sets uri", true, "1", "ipfs://uri-1", nil},
		{"empty uri on minted token is allowed", true, "1", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenLedger, meta, metaLedger := newMeta(0, cur)
			if tt.mint {
				urequire.NoError(t, tokenLedger.Mint(testutils.TestAddress("alice"), tt.tid))
			}

			err := metaLedger.SetTokenURI(tt.tid, tt.uri)
			if tt.wantErr != nil {
				uassert.ErrorIs(t, err, tt.wantErr)
				uassert.False(t, meta.HasTokenURI(tt.tid))
				return
			}

			urequire.NoError(t, err)
			uassert.True(t, meta.HasTokenURI(tt.tid))
			got, err := meta.TokenURI(tt.tid)
			urequire.NoError(t, err)
			uassert.Equal(t, tt.uri, got)
		})
	}
}

func TestSetTokenURIOverwrite(cur realm, t *testing.T) {
	tokenLedger, meta, metaLedger := newMeta(0, cur)
	urequire.NoError(t, tokenLedger.Mint(testutils.TestAddress("alice"), "1"))

	urequire.NoError(t, metaLedger.SetTokenURI("1", "ipfs://first"))
	urequire.NoError(t, metaLedger.SetTokenURI("1", "ipfs://second"))

	got, err := meta.TokenURI("1")
	urequire.NoError(t, err)
	uassert.Equal(t, "ipfs://second", got)
}

func TestSetTokenMetadata(cur realm, t *testing.T) {
	full := Data{
		Image:           "ipfs://img-1",
		ImageData:       "<svg/>",
		ExternalURL:     "https://example.com/1",
		Description:     "first",
		Name:            "Token One",
		Attributes:      []Trait{{DisplayType: "string", TraitType: "rarity", Value: "rare"}},
		BackgroundColor: "ffffff",
		AnimationURL:    "ipfs://anim-1",
		YoutubeURL:      "https://youtu.be/abc",
	}

	tests := []struct {
		name    string
		mint    bool
		tid     grc721.TokenID
		data    Data
		wantErr error
	}{
		{"unminted token returns ErrTokenNotMinted", false, "404", Data{Name: "x"}, ErrTokenNotMinted},
		{"minted token stores full metadata", true, "1", full, nil},
		{"minted token stores zero-value metadata", true, "2", Data{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenLedger, meta, metaLedger := newMeta(0, cur)
			if tt.mint {
				urequire.NoError(t, tokenLedger.Mint(testutils.TestAddress("alice"), tt.tid))
			}

			err := metaLedger.SetTokenMetadata(tt.tid, tt.data)
			if tt.wantErr != nil {
				uassert.ErrorIs(t, err, tt.wantErr)
				uassert.False(t, meta.HasMetadata(tt.tid))
				return
			}

			urequire.NoError(t, err)
			uassert.True(t, meta.HasMetadata(tt.tid))
			got, err := meta.TokenMetadata(tt.tid)
			urequire.NoError(t, err)
			uassert.Equal(t, tt.data.Name, got.Name)
			uassert.Equal(t, tt.data.Description, got.Description)
			uassert.Equal(t, tt.data.Image, got.Image)
			uassert.Equal(t, len(tt.data.Attributes), len(got.Attributes))
			if len(tt.data.Attributes) > 0 {
				uassert.Equal(t, tt.data.Attributes[0].Value, got.Attributes[0].Value)
				uassert.Equal(t, tt.data.Attributes[0].TraitType, got.Attributes[0].TraitType)
			}
		})
	}
}

func TestSetTokenMetadataOverwrite(cur realm, t *testing.T) {
	tokenLedger, meta, metaLedger := newMeta(0, cur)
	urequire.NoError(t, tokenLedger.Mint(testutils.TestAddress("alice"), "1"))

	urequire.NoError(t, metaLedger.SetTokenMetadata("1", Data{Name: "first"}))
	urequire.NoError(t, metaLedger.SetTokenMetadata("1", Data{Name: "second"}))

	got, err := meta.TokenMetadata("1")
	urequire.NoError(t, err)
	uassert.Equal(t, "second", got.Name)
}

func TestReadsOnUnsetToken(cur realm, t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T, meta *Metadata)
	}{
		{"TokenURI on unset token returns ErrNoTokenURI", func(t *testing.T, meta *Metadata) {
			uri, err := meta.TokenURI("1")
			uassert.Equal(t, "", uri)
			uassert.ErrorIs(t, err, ErrNoTokenURI)
		}},
		{"TokenMetadata on unset token returns ErrNoMetadata", func(t *testing.T, meta *Metadata) {
			data, err := meta.TokenMetadata("1")
			uassert.Equal(t, "", data.Name)
			uassert.ErrorIs(t, err, ErrNoMetadata)
		}},
		{"HasTokenURI is false on unset token", func(t *testing.T, meta *Metadata) {
			uassert.False(t, meta.HasTokenURI("1"))
		}},
		{"HasMetadata is false on unset token", func(t *testing.T, meta *Metadata) {
			uassert.False(t, meta.HasMetadata("1"))
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, meta, _ := newMeta(0, cur)
			tt.fn(t, meta)
		})
	}
}

func TestOnBurnClearsMetadata(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	tokenLedger, meta, metaLedger := newMeta(0, cur)

	urequire.NoError(t, tokenLedger.Mint(alice, "1"))
	urequire.NoError(t, metaLedger.SetTokenURI("1", "ipfs://uri-1"))
	urequire.NoError(t, metaLedger.SetTokenMetadata("1", Data{Name: "one"}))
	uassert.True(t, meta.HasTokenURI("1"))
	uassert.True(t, meta.HasMetadata("1"))

	urequire.NoError(t, tokenLedger.Burn("1"))
	uassert.False(t, meta.HasTokenURI("1"))
	uassert.False(t, meta.HasMetadata("1"))

	_, err := meta.TokenURI("1")
	uassert.ErrorIs(t, err, ErrNoTokenURI)
	_, err = meta.TokenMetadata("1")
	uassert.ErrorIs(t, err, ErrNoMetadata)

	urequire.NoError(t, tokenLedger.Mint(alice, "1"))
	uassert.False(t, meta.HasTokenURI("1"))
	uassert.False(t, meta.HasMetadata("1"))
}

func TestHooks(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	tokenLedger, meta, metaLedger := newMeta(0, cur)

	urequire.NoError(t, tokenLedger.Mint(alice, "1"))
	urequire.NoError(t, metaLedger.SetTokenURI("1", "ipfs://uri-1"))
	urequire.NoError(t, metaLedger.SetTokenMetadata("1", Data{Name: "one"}))

	metaLedger.OnMint(alice, "1")
	metaLedger.OnTransfer(alice, bob, "1")
	uassert.True(t, meta.HasTokenURI("1"))
	uassert.True(t, meta.HasMetadata("1"))
	uri, err := meta.TokenURI("1")
	urequire.NoError(t, err)
	uassert.Equal(t, "ipfs://uri-1", uri)

	metaLedger.OnBurn("1")
	uassert.False(t, meta.HasTokenURI("1"))
	uassert.False(t, meta.HasMetadata("1"))
}
