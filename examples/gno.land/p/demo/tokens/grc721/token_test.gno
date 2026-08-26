package grc721

import (
	"strings"
	"testing"

	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

func mustBalanceOf(t *testing.T, tok *Token, addr address) int64 {
	balance, err := tok.BalanceOf(addr)
	urequire.NoError(t, err)
	return balance
}

// newTestToken builds via same-realm cross (NewToken requires rlm.IsCurrent()).
func newTestToken(name, symbol string, id seqid.ID, rlm realm) (tok *Token, led *PrivateLedger) {
	func(cur realm) {
		tok, led = NewToken(name, symbol, id, cur)
	}(cross(rlm))
	return
}

type mockExtension struct {
	mints, transfers, burns int
	lastMintTo              address
	lastTransferFrom        address
	lastTransferTo          address
	lastBurn                TokenID
}

func (m *mockExtension) ExtensionKind() string { return "mock" }

func (m *mockExtension) OnMint(to address, tid TokenID) {
	m.mints++
	m.lastMintTo = to
}

func (m *mockExtension) OnTransfer(from, to address, tid TokenID) {
	m.transfers++
	m.lastTransferFrom = from
	m.lastTransferTo = to
}

func (m *mockExtension) OnBurn(tid TokenID) {
	m.burns++
	m.lastBurn = tid
}

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
	carl  = testutils.TestAddress("carl")
)

func TestNewToken(cur realm, t *testing.T) {
	tok, led := newTestToken("Foo NFT", "FOO", 0, cur)
	urequire.False(t, tok == nil, "token should not be nil")
	urequire.False(t, led == nil, "ledger should not be nil")

	uassert.Equal(t, "Foo NFT", tok.GetName())
	uassert.Equal(t, "FOO", tok.GetSymbol())
	uassert.Equal(t, int64(0), tok.TotalSupply())
	uassert.Equal(t, 0, tok.KnownAccounts())
	uassert.True(t, strings.HasSuffix(tok.ID(), ".FOO."+seqid.ID(0).String()), "ID must end with .symbol.id")
	uassert.True(t, strings.HasPrefix(tok.ID(), "gno.land/"), "ID must begin with the origin realm path")
	uassert.Equal(t, tok.ID(), led.ReadToken().ID())
}

func TestNewTokenDistinctIDs(cur realm, t *testing.T) {
	first, _ := newTestToken("Same", "DUP", 1, cur)
	second, _ := newTestToken("Same", "DUP", 2, cur)
	uassert.True(t, first.ID() != second.ID(), "distinct seqid must yield distinct IDs")

	a, _ := newTestToken("Same", "DUP", 3, cur)
	b, _ := newTestToken("Same", "DUP", 3, cur)
	uassert.Equal(t, a.ID(), b.ID())
}

func TestNewTokenValidation(cur realm, t *testing.T) {
	longName := strings.Repeat("a", MaxNameLen+1)
	maxName := strings.Repeat("a", MaxNameLen)
	longSymbol := strings.Repeat("A", MaxSymbolLen+1)

	tests := []struct {
		name       string
		tokenName  string
		symbol     string
		wantAbort  bool
		wantSubstr string
	}{
		{"valid name and symbol succeeds", "Foo", "FOO", false, ""},
		{"symbol with full slug charset succeeds", "Foo", "aZ9_-", false, ""},
		{"name at max length succeeds", maxName, "FOO", false, ""},
		{"empty name returns ErrInvalidName", "", "FOO", true, "invalid token name"},
		{"name over max length returns ErrInvalidName", longName, "FOO", true, "invalid token name"},
		{"name with low control char returns ErrInvalidName", "Foo\x01", "FOO", true, "invalid token name"},
		{"name with DEL control char returns ErrInvalidName", "Foo\x7f", "FOO", true, "invalid token name"},
		{"empty symbol returns ErrInvalidSymbol", "Foo", "", true, "invalid token symbol"},
		{"symbol over max length returns ErrInvalidSymbol", "Foo", longSymbol, true, "invalid token symbol"},
		{"symbol with space returns ErrInvalidSymbol", "Foo", "has space", true, "invalid token symbol"},
		{"symbol with dot returns ErrInvalidSymbol", "Foo", "FO.O", true, "invalid token symbol"},
		{"symbol with slash returns ErrInvalidSymbol", "Foo", "FO/O", true, "invalid token symbol"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantAbort {
				urequire.AbortsContains(t, cur, tc.wantSubstr, func() {
					newTestToken(tc.tokenName, tc.symbol, 0, cur)
				})
				return
			}
			tok, _ := newTestToken(tc.tokenName, tc.symbol, 0, cur)
			uassert.Equal(t, tc.tokenName, tok.GetName())
			uassert.Equal(t, tc.symbol, tok.GetSymbol())
		})
	}
}

func TestNewTokenRejectsSpoofedRealm(cur realm, t *testing.T) {
	// Stale outer cur in a fresh frame → ErrSpoofedRealm.
	urequire.AbortsContains(t, cur, "rlm does not match the current crossing frame", func() {
		stale := cur
		func(cur realm) {
			NewToken("Foo", "FOO", 0, stale)
		}(cross(cur))
	})
}

func TestMint(cur realm, t *testing.T) {
	tests := []struct {
		name    string
		action  func(led *PrivateLedger) error
		wantErr error
		verify  func(t *testing.T, tok *Token, led *PrivateLedger)
	}{
		{
			name: "mint to valid address succeeds",
			action: func(led *PrivateLedger) error {
				return led.Mint(alice, "1")
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				uassert.Equal(t, int64(1), tok.TotalSupply())
				uassert.Equal(t, int64(1), mustBalanceOf(t, tok, alice))
				uassert.Equal(t, 1, tok.KnownAccounts())
				owner, err := tok.OwnerOf("1")
				urequire.NoError(t, err)
				uassert.Equal(t, alice, owner)
			},
		},
		{
			name: "mint to invalid address returns ErrInvalidAddress",
			action: func(led *PrivateLedger) error {
				return led.Mint(zeroAddress, "1")
			},
			wantErr: ErrInvalidAddress,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				uassert.Equal(t, int64(0), tok.TotalSupply())
			},
		},
		{
			name: "mint duplicate token id returns ErrTokenIdAlreadyExists",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				return led.Mint(bob, "1")
			},
			wantErr: ErrTokenIdAlreadyExists,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				owner, err := tok.OwnerOf("1")
				urequire.NoError(t, err)
				uassert.Equal(t, alice, owner)
				uassert.Equal(t, int64(1), tok.TotalSupply())
			},
		},
		{
			name: "minting distinct ids to same owner accumulates balance",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				return led.Mint(alice, "2")
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				uassert.Equal(t, int64(2), mustBalanceOf(t, tok, alice))
				uassert.Equal(t, int64(2), tok.TotalSupply())
				uassert.Equal(t, 1, tok.KnownAccounts())
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tok, led := newTestToken("Foo", "FOO", 0, cur)
			err := tc.action(led)
			if tc.wantErr == nil {
				uassert.NoError(t, err)
			} else {
				uassert.ErrorIs(t, err, tc.wantErr)
			}
			if tc.verify != nil {
				tc.verify(t, tok, led)
			}
		})
	}
}

func TestBurn(cur realm, t *testing.T) {
	tests := []struct {
		name    string
		action  func(led *PrivateLedger) error
		wantErr error
		verify  func(t *testing.T, tok *Token, led *PrivateLedger)
	}{
		{
			name: "burn existing token clears all state",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				return led.Burn("1")
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				uassert.Equal(t, int64(0), tok.TotalSupply())
				uassert.Equal(t, int64(0), mustBalanceOf(t, tok, alice))
				uassert.Equal(t, 0, tok.KnownAccounts())
				_, err := tok.OwnerOf("1")
				uassert.ErrorIs(t, err, ErrInvalidTokenId)
			},
		},
		{
			name: "burn clears any outstanding approval",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				urequire.NoError(t, led.Approve(alice, bob, "1"))
				return led.Burn("1")
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				uassert.Equal(t, 0, led.tokenApprovals.Size())
				_, err := tok.GetApproved("1")
				uassert.ErrorIs(t, err, ErrInvalidTokenId)
			},
		},
		{
			name: "burn missing token returns ErrInvalidTokenId",
			action: func(led *PrivateLedger) error {
				return led.Burn("999")
			},
			wantErr: ErrInvalidTokenId,
		},
		{
			name: "burn one of two tokens keeps the account known",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				urequire.NoError(t, led.Mint(alice, "2"))
				return led.Burn("1")
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				uassert.Equal(t, int64(1), mustBalanceOf(t, tok, alice))
				uassert.Equal(t, 1, tok.KnownAccounts())
				uassert.Equal(t, int64(1), tok.TotalSupply())
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tok, led := newTestToken("Foo", "FOO", 0, cur)
			err := tc.action(led)
			if tc.wantErr == nil {
				uassert.NoError(t, err)
			} else {
				uassert.ErrorIs(t, err, tc.wantErr)
			}
			if tc.verify != nil {
				tc.verify(t, tok, led)
			}
		})
	}
}

func TestApprove(cur realm, t *testing.T) {
	tests := []struct {
		name    string
		action  func(led *PrivateLedger) error
		wantErr error
		verify  func(t *testing.T, tok *Token, led *PrivateLedger)
	}{
		{
			name: "owner approves another account",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				return led.Approve(alice, bob, "1")
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				approved, err := tok.GetApproved("1")
				urequire.NoError(t, err)
				uassert.Equal(t, bob, approved)
			},
		},
		{
			name: "operator approves on behalf of owner",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				urequire.NoError(t, led.SetApprovalForAll(alice, bob, true))
				return led.Approve(bob, carl, "1")
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				approved, err := tok.GetApproved("1")
				urequire.NoError(t, err)
				uassert.Equal(t, carl, approved)
			},
		},
		{
			name: "approve to zero address revokes an existing approval (EIP-721)",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				urequire.NoError(t, led.Approve(alice, bob, "1"))
				return led.Approve(alice, zeroAddress, "1")
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				_, err := tok.GetApproved("1")
				uassert.ErrorIs(t, err, ErrTokenIdNotApproved)
			},
		},
		{
			name: "revoke with no prior approval is a no-op",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				return led.Approve(alice, zeroAddress, "1")
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				_, err := tok.GetApproved("1")
				uassert.ErrorIs(t, err, ErrTokenIdNotApproved)
			},
		},
		{
			name: "approve on missing token returns ErrInvalidTokenId",
			action: func(led *PrivateLedger) error {
				return led.Approve(alice, bob, "999")
			},
			wantErr: ErrInvalidTokenId,
		},
		{
			name: "approve to current owner returns ErrApprovalToCurrentOwner",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				return led.Approve(alice, alice, "1")
			},
			wantErr: ErrApprovalToCurrentOwner,
		},
		{
			name: "non-owner non-operator approve returns ErrCallerIsNotOwnerOrApproved",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				return led.Approve(bob, carl, "1")
			},
			wantErr: ErrCallerIsNotOwnerOrApproved,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tok, led := newTestToken("Foo", "FOO", 0, cur)
			err := tc.action(led)
			if tc.wantErr == nil {
				uassert.NoError(t, err)
			} else {
				uassert.ErrorIs(t, err, tc.wantErr)
			}
			if tc.verify != nil {
				tc.verify(t, tok, led)
			}
		})
	}
}

func TestSetApprovalForAll(cur realm, t *testing.T) {
	tests := []struct {
		name    string
		action  func(led *PrivateLedger) error
		wantErr error
		verify  func(t *testing.T, tok *Token, led *PrivateLedger)
	}{
		{
			name: "grant operator approval",
			action: func(led *PrivateLedger) error {
				return led.SetApprovalForAll(alice, bob, true)
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				uassert.True(t, tok.IsApprovedForAll(alice, bob))
			},
		},
		{
			name: "revoke operator approval releases the node",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.SetApprovalForAll(alice, bob, true))
				return led.SetApprovalForAll(alice, bob, false)
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				uassert.False(t, tok.IsApprovedForAll(alice, bob))
				uassert.Equal(t, 0, led.operatorApprovals.Size())
			},
		},
		{
			name: "revoke without a prior grant stores nothing",
			action: func(led *PrivateLedger) error {
				return led.SetApprovalForAll(alice, bob, false)
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				uassert.False(t, tok.IsApprovedForAll(alice, bob))
				uassert.Equal(t, 0, led.operatorApprovals.Size())
			},
		},
		{
			name: "invalid operator returns ErrInvalidAddress",
			action: func(led *PrivateLedger) error {
				return led.SetApprovalForAll(alice, zeroAddress, true)
			},
			wantErr: ErrInvalidAddress,
		},
		{
			name: "owner equals operator returns ErrApprovalToCurrentOwner",
			action: func(led *PrivateLedger) error {
				return led.SetApprovalForAll(alice, alice, true)
			},
			wantErr: ErrApprovalToCurrentOwner,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tok, led := newTestToken("Foo", "FOO", 0, cur)
			err := tc.action(led)
			if tc.wantErr == nil {
				uassert.NoError(t, err)
			} else {
				uassert.ErrorIs(t, err, tc.wantErr)
			}
			if tc.verify != nil {
				tc.verify(t, tok, led)
			}
		})
	}
}

func TestTransferFrom(cur realm, t *testing.T) {
	tests := []struct {
		name    string
		action  func(led *PrivateLedger) error
		wantErr error
		verify  func(t *testing.T, tok *Token, led *PrivateLedger)
	}{
		{
			name: "owner transfers own token",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				return led.TransferFrom(alice, alice, bob, "1")
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				owner, err := tok.OwnerOf("1")
				urequire.NoError(t, err)
				uassert.Equal(t, bob, owner)
				uassert.Equal(t, int64(0), mustBalanceOf(t, tok, alice))
				uassert.Equal(t, int64(1), mustBalanceOf(t, tok, bob))
			},
		},
		{
			name: "approved spender transfers and approval is cleared",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				urequire.NoError(t, led.Approve(alice, bob, "1"))
				return led.TransferFrom(bob, alice, carl, "1")
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				owner, err := tok.OwnerOf("1")
				urequire.NoError(t, err)
				uassert.Equal(t, carl, owner)
				_, err = tok.GetApproved("1")
				uassert.ErrorIs(t, err, ErrTokenIdNotApproved)
			},
		},
		{
			name: "operator transfers owner token",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				urequire.NoError(t, led.SetApprovalForAll(alice, bob, true))
				return led.TransferFrom(bob, alice, carl, "1")
			},
			wantErr: nil,
			verify: func(t *testing.T, tok *Token, led *PrivateLedger) {
				owner, err := tok.OwnerOf("1")
				urequire.NoError(t, err)
				uassert.Equal(t, carl, owner)
			},
		},
		{
			name: "unauthorized spender returns ErrCallerIsNotOwnerOrApproved",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				return led.TransferFrom(bob, alice, carl, "1")
			},
			wantErr: ErrCallerIsNotOwnerOrApproved,
		},
		{
			name: "spender on missing token returns ErrCallerIsNotOwnerOrApproved",
			action: func(led *PrivateLedger) error {
				return led.TransferFrom(alice, alice, bob, "999")
			},
			wantErr: ErrCallerIsNotOwnerOrApproved,
		},
		{
			name: "transfer to self returns ErrCannotTransferToSelf",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				return led.TransferFrom(alice, alice, alice, "1")
			},
			wantErr: ErrCannotTransferToSelf,
		},
		{
			name: "transfer from invalid address returns ErrInvalidAddress",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				return led.TransferFrom(alice, zeroAddress, bob, "1")
			},
			wantErr: ErrInvalidAddress,
		},
		{
			name: "transfer to invalid address returns ErrInvalidAddress",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				return led.TransferFrom(alice, alice, zeroAddress, "1")
			},
			wantErr: ErrInvalidAddress,
		},
		{
			name: "transfer from incorrect owner returns ErrTransferFromIncorrectOwner",
			action: func(led *PrivateLedger) error {
				urequire.NoError(t, led.Mint(alice, "1"))
				return led.TransferFrom(alice, bob, carl, "1")
			},
			wantErr: ErrTransferFromIncorrectOwner,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tok, led := newTestToken("Foo", "FOO", 0, cur)
			err := tc.action(led)
			if tc.wantErr == nil {
				uassert.NoError(t, err)
			} else {
				uassert.ErrorIs(t, err, tc.wantErr)
			}
			if tc.verify != nil {
				tc.verify(t, tok, led)
			}
		})
	}
}

func TestReadViews(cur realm, t *testing.T) {
	tok, led := newTestToken("Foo NFT", "FOO", 0, cur)
	urequire.NoError(t, led.Mint(alice, "1"))
	urequire.NoError(t, led.Mint(alice, "2"))
	urequire.NoError(t, led.Mint(bob, "3"))
	urequire.NoError(t, led.Approve(alice, carl, "1"))
	urequire.NoError(t, led.SetApprovalForAll(alice, bob, true))

	tests := []struct {
		name  string
		check func(t *testing.T)
	}{
		{"GetName returns collection name", func(t *testing.T) {
			uassert.Equal(t, "Foo NFT", tok.GetName())
		}},
		{"GetSymbol returns collection symbol", func(t *testing.T) {
			uassert.Equal(t, "FOO", tok.GetSymbol())
		}},
		{"ID carries realm, symbol and seqid", func(t *testing.T) {
			uassert.True(t, strings.HasPrefix(tok.ID(), "gno.land/"))
			uassert.True(t, strings.Contains(tok.ID(), ".FOO."))
		}},
		{"TotalSupply counts live tokens", func(t *testing.T) {
			uassert.Equal(t, int64(3), tok.TotalSupply())
		}},
		{"KnownAccounts counts distinct holders", func(t *testing.T) {
			uassert.Equal(t, 2, tok.KnownAccounts())
		}},
		{"BalanceOf of a holder", func(t *testing.T) {
			uassert.Equal(t, int64(2), mustBalanceOf(t, tok, alice))
		}},
		{"BalanceOf of a non-holder is zero", func(t *testing.T) {
			uassert.Equal(t, int64(0), mustBalanceOf(t, tok, carl))
		}},
		{"BalanceOf of the zero address returns ErrInvalidAddress", func(t *testing.T) {
			_, err := tok.BalanceOf(zeroAddress)
			uassert.ErrorIs(t, err, ErrInvalidAddress)
		}},
		{"OwnerOf existing token", func(t *testing.T) {
			owner, err := tok.OwnerOf("3")
			urequire.NoError(t, err)
			uassert.Equal(t, bob, owner)
		}},
		{"OwnerOf missing token returns ErrInvalidTokenId", func(t *testing.T) {
			_, err := tok.OwnerOf("999")
			uassert.ErrorIs(t, err, ErrInvalidTokenId)
		}},
		{"GetApproved returns the approved account", func(t *testing.T) {
			approved, err := tok.GetApproved("1")
			urequire.NoError(t, err)
			uassert.Equal(t, carl, approved)
		}},
		{"GetApproved on unapproved token returns ErrTokenIdNotApproved", func(t *testing.T) {
			_, err := tok.GetApproved("2")
			uassert.ErrorIs(t, err, ErrTokenIdNotApproved)
		}},
		{"GetApproved on missing token returns ErrInvalidTokenId", func(t *testing.T) {
			_, err := tok.GetApproved("999")
			uassert.ErrorIs(t, err, ErrInvalidTokenId)
		}},
		{"IsApprovedForAll true for operator", func(t *testing.T) {
			uassert.True(t, tok.IsApprovedForAll(alice, bob))
		}},
		{"IsApprovedForAll false for non-operator", func(t *testing.T) {
			uassert.False(t, tok.IsApprovedForAll(alice, carl))
		}},
		{"RenderHome renders header and counters", func(t *testing.T) {
			out := tok.RenderHome()
			uassert.True(t, strings.Contains(out, "# Foo NFT ($FOO)"), "header")
			uassert.True(t, strings.Contains(out, "**Total supply**: 3"), "supply")
			uassert.True(t, strings.Contains(out, "**Known accounts**: 2"), "accounts")
		}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, tc.check)
	}
}

func TestRegisterExtensionNil(cur realm, t *testing.T) {
	_, led := newTestToken("Foo", "FOO", 0, cur)
	uassert.PanicsWithMessage(t, cur, "grc721: nil extension", func() {
		led.RegisterExtension(nil)
	})
}

func TestExtensionHookFanOut(cur realm, t *testing.T) {
	tests := []struct {
		name   string
		run    func(t *testing.T, led *PrivateLedger, ext *mockExtension)
		verify func(t *testing.T, ext *mockExtension)
	}{
		{
			name: "duplicate registration is ignored and mint fires once",
			run: func(t *testing.T, led *PrivateLedger, ext *mockExtension) {
				led.RegisterExtension(ext)
				led.RegisterExtension(ext)
				urequire.NoError(t, led.Mint(alice, "1"))
			},
			verify: func(t *testing.T, ext *mockExtension) {
				uassert.Equal(t, 1, ext.mints)
				uassert.Equal(t, alice, ext.lastMintTo)
			},
		},
		{
			name: "mint, transfer and burn each notify the extension",
			run: func(t *testing.T, led *PrivateLedger, ext *mockExtension) {
				led.RegisterExtension(ext)
				urequire.NoError(t, led.Mint(alice, "1"))
				urequire.NoError(t, led.TransferFrom(alice, alice, bob, "1"))
				urequire.NoError(t, led.Burn("1"))
			},
			verify: func(t *testing.T, ext *mockExtension) {
				uassert.Equal(t, 1, ext.mints)
				uassert.Equal(t, 1, ext.transfers)
				uassert.Equal(t, alice, ext.lastTransferFrom)
				uassert.Equal(t, bob, ext.lastTransferTo)
				uassert.Equal(t, 1, ext.burns)
				uassert.Equal(t, "1", ext.lastBurn.String())
			},
		},
		{
			name: "failed operations do not notify the extension",
			run: func(t *testing.T, led *PrivateLedger, ext *mockExtension) {
				led.RegisterExtension(ext)
				uassert.ErrorIs(t, led.Mint(zeroAddress, "1"), ErrInvalidAddress)
				uassert.ErrorIs(t, led.Burn("999"), ErrInvalidTokenId)
			},
			verify: func(t *testing.T, ext *mockExtension) {
				uassert.Equal(t, 0, ext.mints)
				uassert.Equal(t, 0, ext.burns)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, led := newTestToken("Foo", "FOO", 0, cur)
			ext := &mockExtension{}
			tc.run(t, led, ext)
			tc.verify(t, ext)
		})
	}
}

func TestReadToken(cur realm, t *testing.T) {
	tok, led := newTestToken("Foo", "FOO", 0, cur)
	uassert.Equal(t, tok.ID(), led.ReadToken().ID())
	uassert.True(t, led.ReadToken() == tok, "ReadToken returns the same *Token")
}
