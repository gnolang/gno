package enumerable

import (
	"testing"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// Same-realm cross: NewToken requires rlm.IsCurrent().
func newCore(name, symbol string, id seqid.ID, rlm realm) (tok *grc721.Token, led *grc721.PrivateLedger) {
	func(cur realm) {
		tok, led = grc721.NewToken(name, symbol, id, cur)
	}(cross(rlm))

	return
}

func newEnum(cur realm) (enum *Enumerable, core *grc721.PrivateLedger, led *Ledger) {
	_, coreLedger := newCore("Foo", "FOO", 0, cur)
	enum, led = NewEnumerable(coreLedger)
	return enum, coreLedger, led
}

func globalOrder(t *testing.T, enum *Enumerable) []string {
	t.Helper()
	out := []string{}
	for i := int64(0); i < enum.TotalSupply(); i++ {
		tid, err := enum.TokenByIndex(i)
		urequire.NoError(t, err)
		out = append(out, tid.String())
	}
	return out
}

func ownerOrder(t *testing.T, enum *Enumerable, owner address) []string {
	t.Helper()
	out := []string{}
	for i := int64(0); ; i++ {
		tid, err := enum.TokenOfOwnerByIndex(owner, i)
		if err != nil {
			break
		}
		out = append(out, tid.String())
	}
	return out
}

func eqStrings(t *testing.T, name string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: length got %d want %d (%v vs %v)", name, len(got), len(want), got, want)
		return
	}
	for i := range want {
		uassert.Equal(t, want[i], got[i])
	}
}

func TestNewEnumerable(cur realm, t *testing.T) {
	uassert.PanicsWithMessage(t, cur, "enumerable: nil core ledger", func() {
		NewEnumerable(nil)
	})

	tok, coreLedger := newCore("Foo", "FOO", 0, cur)
	enum, led := NewEnumerable(coreLedger)

	urequire.True(t, enum != nil, "enum built")
	urequire.True(t, led != nil, "ledger built")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ExtensionKind is enumerable", enum.ExtensionKind(), Kind},
		{"ExtensionKind constant value", Kind, "enumerable"},
		{"TokenID matches core token id", enum.TokenID(), tok.ID()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uassert.Equal(t, tt.want, tt.got)
		})
	}

	uassert.Equal(t, int64(0), enum.TotalSupply())
}

func TestReadIndices(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carol := testutils.TestAddress("carol")

	enum, coreLedger, _ := newEnum(cur)
	urequire.NoError(t, coreLedger.Mint(alice, "1"))
	urequire.NoError(t, coreLedger.Mint(alice, "2"))
	urequire.NoError(t, coreLedger.Mint(bob, "3"))

	uassert.Equal(t, int64(3), enum.TotalSupply())

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"TokenByIndex 0 returns first minted", func(t *testing.T) {
			tid, err := enum.TokenByIndex(0)
			urequire.NoError(t, err)
			uassert.Equal(t, "1", tid.String())
		}},
		{"TokenByIndex last returns last minted", func(t *testing.T) {
			tid, err := enum.TokenByIndex(2)
			urequire.NoError(t, err)
			uassert.Equal(t, "3", tid.String())
		}},
		{"TokenByIndex negative returns ErrIndexOutOfRange", func(t *testing.T) {
			_, err := enum.TokenByIndex(-1)
			uassert.ErrorIs(t, err, ErrIndexOutOfRange)
		}},
		{"TokenByIndex out of range returns ErrIndexOutOfRange", func(t *testing.T) {
			_, err := enum.TokenByIndex(3)
			uassert.ErrorIs(t, err, ErrIndexOutOfRange)
		}},
		{"TokenOfOwnerByIndex returns owner's token", func(t *testing.T) {
			tid, err := enum.TokenOfOwnerByIndex(alice, 1)
			urequire.NoError(t, err)
			uassert.Equal(t, "2", tid.String())
		}},
		{"TokenOfOwnerByIndex unknown owner returns ErrIndexOutOfRange", func(t *testing.T) {
			_, err := enum.TokenOfOwnerByIndex(carol, 0)
			uassert.ErrorIs(t, err, ErrIndexOutOfRange)
		}},
		{"TokenOfOwnerByIndex negative index returns ErrIndexOutOfRange", func(t *testing.T) {
			_, err := enum.TokenOfOwnerByIndex(alice, -1)
			uassert.ErrorIs(t, err, ErrIndexOutOfRange)
		}},
		{"TokenOfOwnerByIndex past end returns ErrIndexOutOfRange", func(t *testing.T) {
			_, err := enum.TokenOfOwnerByIndex(alice, 2)
			uassert.ErrorIs(t, err, ErrIndexOutOfRange)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}

	eqStrings(t, "global order", globalOrder(t, enum), []string{"1", "2", "3"})
	eqStrings(t, "alice order", ownerOrder(t, enum, alice), []string{"1", "2"})
	eqStrings(t, "bob order", ownerOrder(t, enum, bob), []string{"3"})
}

func TestBurnGlobalSwapAndPop(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")

	tests := []struct {
		name      string
		burn      grc721.TokenID
		wantOrder []string
	}{
		{"burn last token truncates global list (idx == last)", "3", []string{"1", "2"}},
		{"burn middle token swap-and-pops the global list (idx != last)", "2", []string{"1", "3"}},
		{"burn first token swap-and-pops the global list (idx != last)", "1", []string{"3", "2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enum, coreLedger, _ := newEnum(cur)
			urequire.NoError(t, coreLedger.Mint(alice, "1"))
			urequire.NoError(t, coreLedger.Mint(alice, "2"))
			urequire.NoError(t, coreLedger.Mint(alice, "3"))

			urequire.NoError(t, coreLedger.Burn(tt.burn))

			uassert.Equal(t, int64(2), enum.TotalSupply())
			eqStrings(t, tt.name, globalOrder(t, enum), tt.wantOrder)
		})
	}
}

func TestBurnOwnerSwapAndPop(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")

	tests := []struct {
		name       string
		mint       []grc721.TokenID
		burn       grc721.TokenID
		wantOwner  []string
		ownerEmpty bool
	}{
		{
			name:      "burn first of owner swap-and-pops per-owner list (idx != last)",
			mint:      []grc721.TokenID{"1", "2", "3"},
			burn:      "1",
			wantOwner: []string{"3", "2"},
		},
		{
			name:      "burn middle of owner swap-and-pops per-owner list (idx != last)",
			mint:      []grc721.TokenID{"1", "2", "3"},
			burn:      "2",
			wantOwner: []string{"1", "3"},
		},
		{
			name:      "burn last of owner truncates per-owner list (idx == last)",
			mint:      []grc721.TokenID{"1", "2", "3"},
			burn:      "3",
			wantOwner: []string{"1", "2"},
		},
		{
			name:       "burn only token empties per-owner list (owned.Remove)",
			mint:       []grc721.TokenID{"1"},
			burn:       "1",
			ownerEmpty: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enum, coreLedger, _ := newEnum(cur)
			for _, tid := range tt.mint {
				urequire.NoError(t, coreLedger.Mint(alice, tid))
			}

			urequire.NoError(t, coreLedger.Burn(tt.burn))

			if tt.ownerEmpty {
				_, err := enum.TokenOfOwnerByIndex(alice, 0)
				uassert.ErrorIs(t, err, ErrIndexOutOfRange)
				return
			}
			eqStrings(t, tt.name, ownerOrder(t, enum, alice), tt.wantOwner)
		})
	}
}

func TestTransfer(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	enum, coreLedger, _ := newEnum(cur)
	urequire.NoError(t, coreLedger.Mint(alice, "1"))
	urequire.NoError(t, coreLedger.Mint(alice, "2"))
	urequire.NoError(t, coreLedger.Mint(bob, "3"))

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"transfer 1 alice->bob swap-and-pops alice list", func(t *testing.T) {
			urequire.NoError(t, coreLedger.TransferFrom(alice, alice, bob, "1"))
			eqStrings(t, "alice after transfer", ownerOrder(t, enum, alice), []string{"2"})
			eqStrings(t, "bob after transfer", ownerOrder(t, enum, bob), []string{"3", "1"})
		}},
		{"transfer 2 alice->bob empties alice list", func(t *testing.T) {
			urequire.NoError(t, coreLedger.TransferFrom(alice, alice, bob, "2"))
			_, err := enum.TokenOfOwnerByIndex(alice, 0)
			uassert.ErrorIs(t, err, ErrIndexOutOfRange)
			eqStrings(t, "bob owns all", ownerOrder(t, enum, bob), []string{"3", "1", "2"})
		}},
		{"global supply unchanged by transfers", func(t *testing.T) {
			uassert.Equal(t, int64(3), enum.TotalSupply())
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestDefensiveNoOps(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	enum, coreLedger, led := newEnum(cur)
	urequire.NoError(t, coreLedger.Mint(alice, "1"))

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"removeFromAll unknown id is a no-op", func(t *testing.T) {
			led.removeFromAll("999")
			uassert.Equal(t, int64(1), enum.TotalSupply())
		}},
		{"removeFromOwner unknown owner is a no-op", func(t *testing.T) {
			led.removeFromOwner(bob, "1")
			eqStrings(t, "alice untouched", ownerOrder(t, enum, alice), []string{"1"})
		}},
		{"removeFromOwner known owner unknown id is a no-op", func(t *testing.T) {
			led.removeFromOwner(alice, "999")
			eqStrings(t, "alice untouched", ownerOrder(t, enum, alice), []string{"1"})
		}},
		{"OnBurn of never-minted id is a no-op", func(t *testing.T) {
			led.OnBurn("neverminted")
			uassert.Equal(t, int64(1), enum.TotalSupply())
			eqStrings(t, "alice untouched", ownerOrder(t, enum, alice), []string{"1"})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
