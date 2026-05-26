package grc721

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

const (
	tellerTokenName   = "TellerToken"
	tellerTokenSymbol = "TLR"
)

func TestCallerTellerImpl(cur realm, t *testing.T) {
	tok, _ := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	teller := tok.CallerTeller()
	urequire.True(t, tok != nil)
	var _ Teller = teller
}

// evilWrap embeds Teller; the seal-marker pattern (if we had one) would be
// satisfied via promotion, but IsCanonicalTeller rejects it because *evilWrap
// is nominally distinct from *fnTeller.
type evilWrap struct {
	Teller
}

func TestIsCanonicalTeller(cur realm, t *testing.T) {
	tok, _ := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	legit := tok.CallerTeller()

	uassert.True(t, IsCanonicalTeller(legit),
		"canonical *fnTeller from CallerTeller must pass")

	evil := &evilWrap{Teller: legit}
	uassert.False(t, IsCanonicalTeller(evil),
		"foreign type embedding a canonical Teller must be rejected")
}

func TestCallerTeller_TransferFrom(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))

	teller := tok.CallerTeller()

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		uassert.NoError(t, teller.TransferFrom(0, cur, alice, bob, TokenID("1")))
	}(cross(cur))

	owner, err := tok.OwnerOf(TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, bob.String(), owner.String())
}

func TestCallerTeller_Approve(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))

	teller := tok.CallerTeller()

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		uassert.NoError(t, teller.Approve(0, cur, bob, TokenID("1")))
	}(cross(cur))

	got, err := tok.GetApproved(TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, bob.String(), got.String())
}

func TestCallerTeller_Approve_NotOwner(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	alice := testutils.TestAddress("alice")
	eve := testutils.TestAddress("eve")
	bob := testutils.TestAddress("bob")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))

	teller := tok.CallerTeller()

	testing.SetRealm(testing.NewUserRealm(eve))
	func(cur realm) {
		err := teller.Approve(0, cur, bob, TokenID("1"))
		uassert.ErrorIs(t, err, ErrCallerIsNotOwnerOrApproved)
	}(cross(cur))
}

func TestCallerTeller_TransferFrom_NotApproved(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	eve := testutils.TestAddress("eve")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))

	teller := tok.CallerTeller()

	testing.SetRealm(testing.NewUserRealm(eve))
	func(cur realm) {
		err := teller.TransferFrom(0, cur, alice, bob, TokenID("1"))
		uassert.ErrorIs(t, err, ErrCallerIsNotOwnerOrApproved)
	}(cross(cur))
}

func TestCallerTeller_TransferFrom_ViaApprovalForAll(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	alice := testutils.TestAddress("alice")
	op := testutils.TestAddress("operator")
	bob := testutils.TestAddress("bob")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))
	uassert.NoError(t, ledger.SetApprovalForAll(alice, op, true))

	teller := tok.CallerTeller()

	testing.SetRealm(testing.NewUserRealm(op))
	func(cur realm) {
		uassert.NoError(t, teller.TransferFrom(0, cur, alice, bob, TokenID("1")))
	}(cross(cur))

	owner, _ := tok.OwnerOf(TokenID("1"))
	uassert.Equal(t, bob.String(), owner.String())
}

func TestCallerTeller_SetApprovalForAll(cur realm, t *testing.T) {
	tok, _ := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	alice := testutils.TestAddress("alice")
	op := testutils.TestAddress("operator")

	teller := tok.CallerTeller()

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		uassert.NoError(t, teller.SetApprovalForAll(0, cur, op, true))
	}(cross(cur))

	uassert.True(t, tok.IsApprovedForAll(alice, op))
}

func TestReadonlyTeller_WritesReturnError(cur realm, t *testing.T) {
	tok, _ := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	alice := testutils.TestAddress("alice")

	teller := tok.ReadonlyTeller()

	func(cur realm) {
		uassert.ErrorIs(t, teller.TransferFrom(0, cur, alice, alice, TokenID("1")), ErrReadonly)
		uassert.ErrorIs(t, teller.Approve(0, cur, alice, TokenID("1")), ErrReadonly)
		uassert.ErrorIs(t, teller.SetApprovalForAll(0, cur, alice, true), ErrReadonly)
	}(cross(cur))
}

func TestReadonlyTeller_ReadsPassThrough(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	alice := testutils.TestAddress("alice")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))

	teller := tok.ReadonlyTeller()

	uassert.Equal(t, tellerTokenName, teller.GetName())
	uassert.Equal(t, tellerTokenSymbol, teller.GetSymbol())

	bal, err := teller.BalanceOf(alice)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(1), bal)

	owner, err := teller.OwnerOf(TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, alice.String(), owner.String())
}

func TestImpersonateTeller_ActsAsAddress(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))

	imp := ledger.ImpersonateTeller(alice)

	func(cur realm) {
		uassert.NoError(t, imp.TransferFrom(0, cur, alice, bob, TokenID("1")))
	}(cross(cur))

	owner, _ := tok.OwnerOf(TokenID("1"))
	uassert.Equal(t, bob.String(), owner.String())
}

func TestRealmTeller_Smoke(cur realm, t *testing.T) {
	tok, _ := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)

	var teller Teller
	func(cur realm) {
		teller = tok.RealmTeller(0, cur)
	}(cross(cur))
	uassert.True(t, teller != nil, "teller should not be nil")
}

func TestRealmSubTeller_Smoke(cur realm, t *testing.T) {
	tok, _ := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)

	var teller Teller
	func(cur realm) {
		teller = tok.RealmSubTeller(0, cur, "vault")
	}(cross(cur))
	uassert.True(t, teller != nil, "teller should not be nil")
}

func TestAccountSlugAddr_EmptySlug(cur realm, t *testing.T) {
	addr := testutils.TestAddress("base")
	uassert.Equal(t, addr.String(), accountSlugAddr(addr, "").String())
}

func TestAccountSlugAddr_NonEmptySlug(cur realm, t *testing.T) {
	addr := testutils.TestAddress("base")
	derived := accountSlugAddr(addr, "vault")
	uassert.True(t, derived.String() != addr.String(), "derived address should differ from base")
}

func TestImpersonateTeller_InvalidAddress(cur realm, t *testing.T) {
	_, ledger := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)

	defer func() {
		r := recover()
		uassert.True(t, r != nil, "ImpersonateTeller must panic on empty addr")
		uassert.True(t, r == ErrInvalidAddress,
			"ImpersonateTeller must panic with ErrInvalidAddress")
	}()
	ledger.ImpersonateTeller(address(""))
}

func TestCallerTeller_Transfer_ToInvalidAddress(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	alice := testutils.TestAddress("alice")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))

	teller := tok.CallerTeller()

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		err := teller.TransferFrom(0, cur, alice, address(""), TokenID("1"))
		uassert.ErrorIs(t, err, ErrInvalidAddress)
	}(cross(cur))
}

func TestCallerTeller_TransferFrom_WrongFrom(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))
	uassert.NoError(t, ledger.Approve(bob, TokenID("1")))

	teller := tok.CallerTeller()

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		err := teller.TransferFrom(0, cur, bob, alice, TokenID("1"))
		uassert.ErrorIs(t, err, ErrTransferFromIncorrectOwner)
	}(cross(cur))
}

func TestCallerTeller_Approve_ViaApprovalForAll(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	alice := testutils.TestAddress("alice")
	op := testutils.TestAddress("operator")
	bob := testutils.TestAddress("bob")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))
	uassert.NoError(t, ledger.SetApprovalForAll(alice, op, true))

	teller := tok.CallerTeller()

	testing.SetRealm(testing.NewUserRealm(op))
	func(cur realm) {
		uassert.NoError(t, teller.Approve(0, cur, bob, TokenID("1")))
	}(cross(cur))

	got, err := tok.GetApproved(TokenID("1"))
	uassert.NoError(t, err)
	uassert.Equal(t, bob.String(), got.String())
}

func TestCallerTeller_TransferFrom_ResetsApprovalEndToEnd(cur realm, t *testing.T) {
	tok, ledger := newTestToken(0, cur, tellerTokenName, tellerTokenSymbol)
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carl := testutils.TestAddress("carl")

	uassert.NoError(t, ledger.Mint(alice, TokenID("1")))
	uassert.NoError(t, ledger.Approve(bob, TokenID("1")))

	teller := tok.CallerTeller()

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		uassert.NoError(t, teller.TransferFrom(0, cur, alice, carl, TokenID("1")))
	}(cross(cur))

	_, err := tok.GetApproved(TokenID("1"))
	uassert.ErrorIs(t, err, ErrTokenIdNotHasApproved)
}
