package grc721

import (
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

const (
	newTokenName   = "NewToken"
	newTokenSymbol = "NEW"
)

// newTestToken constructs a Token. The IIFE same-realm cross
// promotes the test's EOA-origin cur into a fresh /p/grc721
// CodeRealm cur so NewToken's IsCurrent + non-empty rlm.PkgPath()
// checks pass. cross(rlm) is the explicit-rlm form of bare `cross`.
func newTestToken(_ int, rlm realm, name, symbol string) (tok *Token, adm *PrivateLedger) {
	func(cur realm) {
		tok, adm = NewToken(0, cur, name, symbol)
	}(cross(rlm))
	return
}

func TestNewToken(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	urequire.True(t, tok != nil, "token should not be nil")
	urequire.True(t, ledger != nil, "ledger should not be nil")
	urequire.True(t, ledger.token == tok, "ledger should reference back to token")
}

func TestNewTokenValidation(cur realm, t *testing.T) {
	// Each case wraps NewToken in a crossing closure (same as newTestToken)
	// to satisfy IsCurrent + non-empty pkgPath. The validation under test
	// is the name/symbol checks that run after.
	mustPanic := func(name, sym string, want error, label string) {
		t.Helper()
		r := revive(func() {
			newTestToken(0, cur, name, sym)
		})
		if r == nil {
			t.Errorf("%s: expected panic, got none", label)
			return
		}
		if r != want {
			t.Errorf("%s: expected %v, got %v", label, want, r)
		}
	}

	// Empty name / symbol.
	mustPanic("", "OK", ErrInvalidName, "empty name")
	mustPanic("Name", "", ErrInvalidSymbol, "empty symbol")

	// Length caps.
	mustPanic(strings.Repeat("a", MaxNameLen+1), "OK", ErrInvalidName, "name too long")
	mustPanic("Name", strings.Repeat("A", MaxSymbolLen+1), ErrInvalidSymbol, "symbol too long")

	// Name control characters.
	mustPanic("bad\x01name", "OK", ErrInvalidName, "name with control char")
	mustPanic("bad\nname", "OK", ErrInvalidName, "name with newline")

	// Symbol charset — disallowed delimiters and whitespace.
	mustPanic("Name", "BA.D", ErrInvalidSymbol, "symbol with dot")
	mustPanic("Name", "BA/D", ErrInvalidSymbol, "symbol with slash")
	mustPanic("Name", "BA D", ErrInvalidSymbol, "symbol with space")
	mustPanic("Name", "BA\"D", ErrInvalidSymbol, "symbol with quote")

	// Boundary positives — should NOT panic.
	tok, _ := newTestToken(0, cur, strings.Repeat("a", MaxNameLen), strings.Repeat("A", MaxSymbolLen))
	urequire.True(t, tok != nil, "boundary name+symbol should succeed")

	// UTF-8 name with non-ASCII is allowed.
	tok2, _ := newTestToken(0, cur, "Доллар", "RUB")
	urequire.True(t, tok2 != nil, "UTF-8 name should be allowed")
}

func TestToken_Metadata(cur realm, t *testing.T) {
	tok, _ := newTestToken(0, cur, newTokenName, newTokenSymbol)
	uassert.Equal(t, newTokenName, tok.GetName())
	uassert.Equal(t, newTokenSymbol, tok.GetSymbol())
}

func TestToken_TotalSupply_Empty(cur realm, t *testing.T) {
	tok, _ := newTestToken(0, cur, newTokenName, newTokenSymbol)
	uassert.Equal(t, int64(0), tok.TotalSupply())
}

func TestToken_TotalSupply_AfterMint(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	alice := testutils.TestAddress("alice")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))
	uassert.NoError(t, ledger.Mint(alice, TokenID("2")))

	uassert.Equal(t, int64(2), tok.TotalSupply())
}

func TestToken_BalanceOf(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	bal, err := tok.BalanceOf(alice)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(0), bal)

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))
	uassert.NoError(t, ledger.Mint(alice, TokenID("2")))
	uassert.NoError(t, ledger.Mint(bob, TokenID("3")))

	balAlice, err := tok.BalanceOf(alice)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(2), balAlice)

	balBob, err := tok.BalanceOf(bob)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(1), balBob)
}

func TestToken_BalanceOf_InvalidAddress(cur realm, t *testing.T) {
	tok, _ := newTestToken(0, cur, newTokenName, newTokenSymbol)
	_, err := tok.BalanceOf(address(""))
	uassert.ErrorIs(t, err, ErrInvalidAddress)
}

func TestToken_OwnerOf(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	alice := testutils.TestAddress("alice")

	_, err := tok.OwnerOf(TokenID("missing"))
	uassert.ErrorIs(t, err, ErrInvalidTokenId)

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))
	owner, err := tok.OwnerOf(TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, alice.String(), owner.String())
}

func TestToken_GetApproved(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	owner := testutils.TestAddress("owner")
	approved := testutils.TestAddress("approved")

	_, err := tok.GetApproved(TokenID("1"))
	uassert.ErrorIs(t, err, ErrTokenIdNotHasApproved)

	uassert.NoError(t, ledger.Mint(owner, TokenID("1")))
	uassert.NoError(t, ledger.Approve(approved, TokenID("1")))

	got, err := tok.GetApproved(TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, approved.String(), got.String())
}

func TestToken_IsApprovedForAll(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	owner := testutils.TestAddress("owner")
	op := testutils.TestAddress("operator")

	uassert.False(t, tok.IsApprovedForAll(owner, op))

	uassert.NoError(t, ledger.SetApprovalForAll(owner, op, true))
	uassert.True(t, tok.IsApprovedForAll(owner, op))

	uassert.NoError(t, ledger.SetApprovalForAll(owner, op, false))
	uassert.False(t, tok.IsApprovedForAll(owner, op))
}

func TestToken_ID(cur realm, t *testing.T) {
	tok, _ := newTestToken(0, cur, newTokenName, newTokenSymbol)
	id := tok.ID()
	suffix := "." + newTokenSymbol
	if len(id) < len(suffix) || id[len(id)-len(suffix):] != suffix {
		t.Errorf("ID() = %q; want suffix %q", id, suffix)
	}
}

func TestPrivateLedger_Mint_Duplicate(cur realm, t *testing.T) {
	_, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	alice := testutils.TestAddress("alice")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))
	err := ledger.Mint(alice, TokenID("1"))
	uassert.ErrorIs(t, err, ErrTokenIdAlreadyExists)
}

func TestPrivateLedger_Mint_InvalidAddress(cur realm, t *testing.T) {
	_, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	err := ledger.Mint(address(""), TokenID("1"))
	uassert.ErrorIs(t, err, ErrInvalidAddress)
}

func TestPrivateLedger_Burn(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	alice := testutils.TestAddress("alice")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))
	uassert.NoError(t, ledger.Burn(TokenID("1")))

	_, err := tok.OwnerOf(TokenID("1"))
	uassert.ErrorIs(t, err, ErrInvalidTokenId)
}

func TestPrivateLedger_Burn_Missing(cur realm, t *testing.T) {
	_, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	err := ledger.Burn(TokenID("ghost"))
	uassert.ErrorIs(t, err, ErrInvalidTokenId)
}

func TestPrivateLedger_Transfer_Admin(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	from := testutils.TestAddress("from")
	to := testutils.TestAddress("to")

	uassert.NoError(t, ledger.Mint(from, TokenID("1")))
	uassert.NoError(t, ledger.Transfer(from, to, TokenID("1")))

	bFrom, err := tok.BalanceOf(from)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(0), bFrom)

	bTo, err := tok.BalanceOf(to)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(1), bTo)

	owner, err := tok.OwnerOf(TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, to.String(), owner.String())
}

func TestPrivateLedger_Transfer_WrongFrom(cur realm, t *testing.T) {
	_, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	owner := testutils.TestAddress("owner")
	other := testutils.TestAddress("other")

	uassert.NoError(t, ledger.Mint(owner, TokenID("1")))
	err := ledger.Transfer(other, owner, TokenID("1"))
	uassert.ErrorIs(t, err, ErrTransferFromIncorrectOwner)
}

func TestPrivateLedger_Transfer_SelfRejected(cur realm, t *testing.T) {
	_, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	alice := testutils.TestAddress("alice")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))
	err := ledger.Transfer(alice, alice, TokenID("1"))
	uassert.ErrorIs(t, err, ErrCannotTransferToSelf)
}

func TestPrivateLedger_Approve_Admin(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	owner := testutils.TestAddress("owner")
	approved := testutils.TestAddress("approved")

	uassert.NoError(t, ledger.Mint(owner, TokenID("1")))
	uassert.NoError(t, ledger.Approve(approved, TokenID("1")))

	got, err := tok.GetApproved(TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, approved.String(), got.String())
}

func TestPrivateLedger_Approve_ToOwnerRejected(cur realm, t *testing.T) {
	_, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	owner := testutils.TestAddress("owner")

	uassert.NoError(t, ledger.Mint(owner, TokenID("1")))
	err := ledger.Approve(owner, TokenID("1"))
	uassert.ErrorIs(t, err, ErrApprovalToCurrentOwner)
}

func TestPrivateLedger_SetApprovalForAll(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	owner := testutils.TestAddress("owner")
	op := testutils.TestAddress("operator")

	uassert.NoError(t, ledger.SetApprovalForAll(owner, op, true))
	uassert.True(t, tok.IsApprovedForAll(owner, op))

	uassert.NoError(t, ledger.SetApprovalForAll(owner, op, false))
	uassert.False(t, tok.IsApprovedForAll(owner, op))
}

func TestPrivateLedger_SetApprovalForAll_SelfRejected(cur realm, t *testing.T) {
	_, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	owner := testutils.TestAddress("owner")

	err := ledger.SetApprovalForAll(owner, owner, true)
	uassert.ErrorIs(t, err, ErrSelfApproval)
}

func TestPrivateLedger_SetApprovalForAll_InvalidAddress(cur realm, t *testing.T) {
	_, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	op := testutils.TestAddress("op")

	err := ledger.SetApprovalForAll(address(""), op, true)
	uassert.ErrorIs(t, err, ErrInvalidAddress)

	err = ledger.SetApprovalForAll(op, address(""), true)
	uassert.ErrorIs(t, err, ErrInvalidAddress)
}

func TestPrivateLedger_Transfer_InvalidAddress(cur realm, t *testing.T) {
	_, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	owner := testutils.TestAddress("owner")

	uassert.NoError(t, ledger.Mint(owner, TokenID("1")))

	err := ledger.Transfer(address(""), owner, TokenID("1"))
	uassert.ErrorIs(t, err, ErrInvalidAddress)

	err = ledger.Transfer(owner, address(""), TokenID("1"))
	uassert.ErrorIs(t, err, ErrInvalidAddress)
}

func TestPrivateLedger_Transfer_MissingToken(cur realm, t *testing.T) {
	_, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	a := testutils.TestAddress("a")
	b := testutils.TestAddress("b")

	err := ledger.Transfer(a, b, TokenID("ghost"))
	uassert.ErrorIs(t, err, ErrInvalidTokenId)
}

func TestPrivateLedger_Approve_InvalidAddress(cur realm, t *testing.T) {
	_, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	owner := testutils.TestAddress("owner")

	uassert.NoError(t, ledger.Mint(owner, TokenID("1")))

	err := ledger.Approve(address(""), TokenID("1"))
	uassert.ErrorIs(t, err, ErrInvalidAddress)
}

func TestPrivateLedger_Approve_MissingToken(cur realm, t *testing.T) {
	_, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	to := testutils.TestAddress("to")

	err := ledger.Approve(to, TokenID("ghost"))
	uassert.ErrorIs(t, err, ErrInvalidTokenId)
}

func TestPrivateLedger_Transfer_ResetsApproval(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	owner := testutils.TestAddress("owner")
	approved := testutils.TestAddress("approved")
	to := testutils.TestAddress("to")

	uassert.NoError(t, ledger.Mint(owner, TokenID("1")))
	uassert.NoError(t, ledger.Approve(approved, TokenID("1")))

	got, err := tok.GetApproved(TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, approved.String(), got.String())

	uassert.NoError(t, ledger.Transfer(owner, to, TokenID("1")))

	_, err = tok.GetApproved(TokenID("1"))
	uassert.ErrorIs(t, err, ErrTokenIdNotHasApproved)
}

func TestPrivateLedger_Burn_ClearsAllState(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	owner := testutils.TestAddress("owner")
	approved := testutils.TestAddress("approved")
	tid := TokenID("1")

	uassert.NoError(t, ledger.Mint(owner, tid))
	uassert.NoError(t, ledger.Approve(approved, tid))

	uassert.NoError(t, ledger.Burn(tid))

	_, err := tok.OwnerOf(tid)
	uassert.ErrorIs(t, err, ErrInvalidTokenId)

	_, err = tok.GetApproved(tid)
	uassert.ErrorIs(t, err, ErrTokenIdNotHasApproved)

	bal, err := tok.BalanceOf(owner)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(0), bal)
}

func TestPrivateLedger_Burn_BalanceDecreasesPerToken(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	owner := testutils.TestAddress("owner")

	uassert.NoError(t, ledger.Mint(owner, TokenID("1")))
	uassert.NoError(t, ledger.Mint(owner, TokenID("2")))
	uassert.NoError(t, ledger.Mint(owner, TokenID("3")))

	bal, _ := tok.BalanceOf(owner)
	uassert.Equal(t, int64(3), bal)

	uassert.NoError(t, ledger.Burn(TokenID("2")))

	bal, _ = tok.BalanceOf(owner)
	uassert.Equal(t, int64(2), bal)

	uassert.Equal(t, int64(2), tok.TotalSupply())
}

func TestSafeMint_AliasesMint(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	owner := testutils.TestAddress("owner")

	uassert.NoError(t, ledger.SafeMint(owner, TokenID("1")))

	got, err := tok.OwnerOf(TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, owner.String(), got.String())
}

func TestToken_KnownAccounts(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, newTokenName, newTokenSymbol)
	a := testutils.TestAddress("a")
	b := testutils.TestAddress("b")

	uassert.Equal(t, 0, tok.KnownAccounts())

	uassert.NoError(t, ledger.Mint(a, TokenID("1")))
	uassert.Equal(t, 1, tok.KnownAccounts())

	uassert.NoError(t, ledger.Mint(a, TokenID("2")))
	uassert.Equal(t, 1, tok.KnownAccounts())

	uassert.NoError(t, ledger.Mint(b, TokenID("3")))
	uassert.Equal(t, 2, tok.KnownAccounts())
}
