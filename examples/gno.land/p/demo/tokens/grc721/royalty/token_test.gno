package royalty

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

func newRoy(maxBps int64, rlm realm) (*grc721.PrivateLedger, *Royalty, *Ledger) {
	_, tokenLedger := newToken("Foo", "FOO", 0, rlm)
	roy, royLedger := NewRoyalty(tokenLedger, maxBps)
	return tokenLedger, roy, royLedger
}

func TestNewRoyalty(cur realm, t *testing.T) {
	tests := []struct {
		name      string
		nilLedger bool
		maxBps    int64
		wantPanic string // "" = success
	}{
		{"nil token ledger panics", true, 500, "royalty: nil core ledger"},
		{"maxBps above FeeDenominator panics", false, 10001, ErrMaxBpsRange.Error()},
		{"negative maxBps panics", false, -1, ErrMaxBpsRange.Error()},
		{"maxBps at FeeDenominator boundary constructs", false, 10000, ""},
		{"maxBps zero boundary constructs", false, 0, ""},
		{"typical maxBps constructs", false, 1000, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.nilLedger {
				uassert.PanicsWithMessage(t, cur, tt.wantPanic, func() {
					NewRoyalty(nil, tt.maxBps)
				})
				return
			}

			_, tokenLedger := newToken("Foo", "FOO", 0, cur)
			if tt.wantPanic != "" {
				uassert.PanicsWithMessage(t, cur, tt.wantPanic, func() {
					NewRoyalty(tokenLedger, tt.maxBps)
				})
				return
			}

			roy, royLedger := NewRoyalty(tokenLedger, tt.maxBps)
			uassert.True(t, roy != nil)
			uassert.True(t, royLedger != nil)
			uassert.Equal(t, Kind, roy.ExtensionKind())
			uassert.Equal(t, "royalty", roy.ExtensionKind())
			uassert.Equal(t, tt.maxBps, roy.MaxBps())
			uassert.Equal(t, tokenLedger.ReadToken().ID(), roy.TokenID())
		})
	}
}

func TestRoyaltyInfo(cur realm, t *testing.T) {
	artist := testutils.TestAddress("artist")
	special := testutils.TestAddress("special")

	tests := []struct {
		name         string
		setDefault   bool
		defaultBps   int64
		setToken     bool
		tokenBps     int64
		salePrice    int64
		wantReceiver address
		wantAmount   int64
	}{
		{
			name:         "no royalty configured returns zero receiver and amount",
			salePrice:    10000,
			wantReceiver: zeroAddress,
			wantAmount:   0,
		},
		{
			name:         "default 5% yields 500 on salePrice 10000",
			setDefault:   true,
			defaultBps:   500,
			salePrice:    10000,
			wantReceiver: artist,
			wantAmount:   500,
		},
		{
			name:         "per-token override takes precedence over default",
			setDefault:   true,
			defaultBps:   500,
			setToken:     true,
			tokenBps:     250,
			salePrice:    10000,
			wantReceiver: special,
			wantAmount:   250,
		},
		{
			name:         "fractional bps 250 truncates on small sale price",
			setToken:     true,
			tokenBps:     250,
			salePrice:    100, // 100*250/10000 = 2
			wantReceiver: special,
			wantAmount:   2,
		},
		{
			name:         "zero sale price yields zero amount",
			setDefault:   true,
			defaultBps:   500,
			salePrice:    0,
			wantReceiver: artist,
			wantAmount:   0,
		},
		{
			name:         "default receiver applies when token has no override",
			setDefault:   true,
			defaultBps:   1000,
			salePrice:    5000,
			wantReceiver: artist,
			wantAmount:   500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenLedger, roy, royLedger := newRoy(1000, cur)
			urequire.NoError(t, tokenLedger.Mint(testutils.TestAddress("alice"), "1"))

			if tt.setDefault {
				urequire.NoError(t, royLedger.SetDefaultRoyalty(artist, tt.defaultBps))
			}
			if tt.setToken {
				urequire.NoError(t, royLedger.SetTokenRoyalty("1", special, tt.tokenBps))
			}

			recv, amt, err := roy.RoyaltyInfo("1", tt.salePrice)
			urequire.NoError(t, err)
			uassert.Equal(t, tt.wantReceiver, recv)
			uassert.Equal(t, tt.wantAmount, amt)
		})
	}
}

func TestDefaultRoyalty(cur realm, t *testing.T) {
	artist := testutils.TestAddress("artist")

	t.Run("unset default reports not present", func(t *testing.T) {
		_, roy, _ := newRoy(1000, cur)
		info, ok := roy.DefaultRoyalty()
		uassert.False(t, ok)
		uassert.Equal(t, zeroAddress, info.Receiver)
		uassert.Equal(t, int64(0), info.Bps)
	})

	t.Run("set default is reported and readable", func(t *testing.T) {
		_, roy, royLedger := newRoy(1000, cur)
		urequire.NoError(t, royLedger.SetDefaultRoyalty(artist, 500))
		info, ok := roy.DefaultRoyalty()
		uassert.True(t, ok)
		uassert.Equal(t, artist, info.Receiver)
		uassert.Equal(t, int64(500), info.Bps)
	})

	t.Run("delete default clears it", func(t *testing.T) {
		_, roy, royLedger := newRoy(1000, cur)
		urequire.NoError(t, royLedger.SetDefaultRoyalty(artist, 500))
		royLedger.DeleteDefaultRoyalty()
		info, ok := roy.DefaultRoyalty()
		uassert.False(t, ok)
		uassert.Equal(t, zeroAddress, info.Receiver)

		recv, amt, err := roy.RoyaltyInfo("1", 10000)
		urequire.NoError(t, err)
		uassert.Equal(t, zeroAddress, recv)
		uassert.Equal(t, int64(0), amt)
	})
}

func TestTokenRoyalty(cur realm, t *testing.T) {
	artist := testutils.TestAddress("artist")
	special := testutils.TestAddress("special")

	t.Run("unset per-token override reports not present", func(t *testing.T) {
		_, roy, _ := newRoy(1000, cur)
		info, ok := roy.TokenRoyalty("1")
		uassert.False(t, ok)
		uassert.Equal(t, zeroAddress, info.Receiver)
	})

	t.Run("set per-token override is reported and readable", func(t *testing.T) {
		tokenLedger, roy, royLedger := newRoy(1000, cur)
		urequire.NoError(t, tokenLedger.Mint(testutils.TestAddress("alice"), "1"))
		urequire.NoError(t, royLedger.SetTokenRoyalty("1", special, 250))
		info, ok := roy.TokenRoyalty("1")
		uassert.True(t, ok)
		uassert.Equal(t, special, info.Receiver)
		uassert.Equal(t, int64(250), info.Bps)
	})

	t.Run("delete per-token override falls back to default", func(t *testing.T) {
		tokenLedger, roy, royLedger := newRoy(1000, cur)
		urequire.NoError(t, tokenLedger.Mint(testutils.TestAddress("alice"), "1"))
		urequire.NoError(t, royLedger.SetDefaultRoyalty(artist, 500))
		urequire.NoError(t, royLedger.SetTokenRoyalty("1", special, 250))

		royLedger.DeleteTokenRoyalty("1")
		_, ok := roy.TokenRoyalty("1")
		uassert.False(t, ok)

		recv, amt, err := roy.RoyaltyInfo("1", 10000)
		urequire.NoError(t, err)
		uassert.Equal(t, artist, recv)
		uassert.Equal(t, int64(500), amt)
	})
}

func TestSetDefaultRoyaltyValidation(cur realm, t *testing.T) {
	artist := testutils.TestAddress("artist")

	tests := []struct {
		name     string
		receiver address
		bps      int64
		wantErr  error
	}{
		{"bps above maxBps returns ErrInvalidBps", artist, 1001, ErrInvalidBps},
		{"negative bps returns ErrInvalidBps", artist, -1, ErrInvalidBps},
		{"invalid (zero) receiver returns ErrInvalidReceiver", zeroAddress, 100, ErrInvalidReceiver},
		{"bps at maxBps boundary succeeds", artist, 1000, nil},
		{"zero bps succeeds", artist, 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, roy, royLedger := newRoy(1000, cur)
			err := royLedger.SetDefaultRoyalty(tt.receiver, tt.bps)
			if tt.wantErr != nil {
				uassert.ErrorIs(t, err, tt.wantErr)
				_, ok := roy.DefaultRoyalty()
				uassert.False(t, ok)
				return
			}
			urequire.NoError(t, err)
			info, ok := roy.DefaultRoyalty()
			uassert.True(t, ok)
			uassert.Equal(t, tt.bps, info.Bps)
		})
	}
}

func TestSetTokenRoyaltyValidation(cur realm, t *testing.T) {
	artist := testutils.TestAddress("artist")

	tests := []struct {
		name     string
		mint     bool
		tid      grc721.TokenID
		receiver address
		bps      int64
		wantErr  error
	}{
		{"unminted token returns ErrTokenNotMinted", false, "404", artist, 100, ErrTokenNotMinted},
		{"minted token with bps above maxBps returns ErrInvalidBps", true, "1", artist, 1001, ErrInvalidBps},
		{"minted token with negative bps returns ErrInvalidBps", true, "1", artist, -1, ErrInvalidBps},
		{"minted token with invalid receiver returns ErrInvalidReceiver", true, "1", zeroAddress, 100, ErrInvalidReceiver},
		{"minted token at maxBps boundary succeeds", true, "1", artist, 1000, nil},
		{"minted token with valid override succeeds", true, "1", artist, 250, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenLedger, roy, royLedger := newRoy(1000, cur)
			if tt.mint {
				urequire.NoError(t, tokenLedger.Mint(testutils.TestAddress("alice"), tt.tid))
			}

			err := royLedger.SetTokenRoyalty(tt.tid, tt.receiver, tt.bps)
			if tt.wantErr != nil {
				uassert.ErrorIs(t, err, tt.wantErr)
				_, ok := roy.TokenRoyalty(tt.tid)
				uassert.False(t, ok)
				return
			}
			urequire.NoError(t, err)
			info, ok := roy.TokenRoyalty(tt.tid)
			uassert.True(t, ok)
			uassert.Equal(t, tt.bps, info.Bps)
		})
	}
}

func TestOnBurnClearsTokenRoyaltyKeepsDefault(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	artist := testutils.TestAddress("artist")
	special := testutils.TestAddress("special")
	tokenLedger, roy, royLedger := newRoy(1000, cur)

	urequire.NoError(t, tokenLedger.Mint(alice, "1"))
	urequire.NoError(t, royLedger.SetDefaultRoyalty(artist, 500))
	urequire.NoError(t, royLedger.SetTokenRoyalty("1", special, 250))

	_, ok := roy.TokenRoyalty("1")
	uassert.True(t, ok)

	urequire.NoError(t, tokenLedger.Burn("1"))
	_, ok = roy.TokenRoyalty("1")
	uassert.False(t, ok)

	info, ok := roy.DefaultRoyalty()
	uassert.True(t, ok)
	uassert.Equal(t, artist, info.Receiver)

	urequire.NoError(t, tokenLedger.Mint(alice, "1"))
	_, ok = roy.TokenRoyalty("1")
	uassert.False(t, ok)
	recv, amt, err := roy.RoyaltyInfo("1", 10000)
	urequire.NoError(t, err)
	uassert.Equal(t, artist, recv)
	uassert.Equal(t, int64(500), amt)
}

func TestHooks(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	artist := testutils.TestAddress("artist")
	special := testutils.TestAddress("special")
	tokenLedger, roy, royLedger := newRoy(1000, cur)

	urequire.NoError(t, tokenLedger.Mint(alice, "1"))
	urequire.NoError(t, royLedger.SetDefaultRoyalty(artist, 500))
	urequire.NoError(t, royLedger.SetTokenRoyalty("1", special, 250))

	royLedger.OnMint(alice, "1")
	royLedger.OnTransfer(alice, bob, "1")
	info, ok := roy.TokenRoyalty("1")
	uassert.True(t, ok)
	uassert.Equal(t, special, info.Receiver)
	_, ok = roy.DefaultRoyalty()
	uassert.True(t, ok)

	royLedger.OnBurn("1")
	_, ok = roy.TokenRoyalty("1")
	uassert.False(t, ok)
	_, ok = roy.DefaultRoyalty()
	uassert.True(t, ok)
}
