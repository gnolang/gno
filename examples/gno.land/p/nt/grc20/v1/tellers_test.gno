package grc20

import (
	"math"
	"strconv"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestCallerTellerImpl(cur realm, t *testing.T) {
	tok, ledger := newTestToken("Dummy", "DUMMY", 4, 0, cur)
	teller := ledger.CallerTeller()
	urequire.False(t, tok == nil)
	var _ Teller = teller
}

// The sub-realm case (guardHome must not lock out the token's own realm when it
// operates under an identity minted by its own cur.Sub()) needs a real realm
// frame, since cur.Sub() refuses to mint from a p/ package. It lives in
// filetests/caller_teller_sub_realm_filetest.gno.

// evilWrap embeds Teller; the seal-marker pattern (if we had one) would be
// satisfied via promotion, but IsCanonicalTeller rejects it because *evilWrap
// is nominally distinct from *fnTeller.
type evilWrap struct {
	Teller
}

func TestIsCanonicalTeller(cur realm, t *testing.T) {
	_, ledger := newTestToken("Dummy", "DUMMY", 4, 0, cur)
	legit := ledger.CallerTeller()

	uassert.True(t, IsCanonicalTeller(legit),
		"canonical *fnTeller from CallerTeller must pass")

	evil := &evilWrap{Teller: legit}
	uassert.False(t, IsCanonicalTeller(evil),
		"foreign type embedding a canonical Teller must be rejected")
}

func TestTeller(cur realm, t *testing.T) {
	var (
		alice = testutils.TestAddress("alice")
		bob   = testutils.TestAddress("bob")
		carl  = testutils.TestAddress("carl")
	)

	token, ledger := newTestToken("Dummy", "DUMMY", 6, 0, cur)

	checkBalances := func(aliceEB, bobEB, carlEB uint256) {
		t.Helper()
		exp := fmtBalances(aliceEB, bobEB, carlEB)
		got := fmtBalances(token.BalanceOf(alice), token.BalanceOf(bob), token.BalanceOf(carl))
		uassert.Equal(t, got, exp, "invalid balances")
	}
	checkAllowances := func(abEB, acEB, baEB, bcEB, caEB, cbEB uint256) {
		t.Helper()
		exp := fmtAllowances(abEB, acEB, baEB, bcEB, caEB, cbEB)
		got := fmtAllowances(
			token.Allowance(alice, bob), token.Allowance(alice, carl),
			token.Allowance(bob, alice), token.Allowance(bob, carl),
			token.Allowance(carl, alice), token.Allowance(carl, bob),
		)
		uassert.Equal(t, got, exp, "invalid allowances")
	}

	checkBalances(0, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0)

	urequire.NoError(t, ledger.Mint(alice, 1000))
	urequire.NoError(t, ledger.Mint(alice, 100))
	checkBalances(1100, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0)

	urequire.NoError(t, ledger.Approve(alice, bob, 99999999))
	checkBalances(1100, 0, 0)
	checkAllowances(99999999, 0, 0, 0, 0, 0)

	urequire.NoError(t, ledger.Approve(alice, bob, 400))
	checkBalances(1100, 0, 0)
	checkAllowances(400, 0, 0, 0, 0, 0)

	urequire.Error(t, ledger.TransferFrom(alice, bob, carl, 100000000))
	checkBalances(1100, 0, 0)
	checkAllowances(400, 0, 0, 0, 0, 0)

	urequire.NoError(t, ledger.TransferFrom(alice, bob, carl, 100))
	checkBalances(1000, 0, 100)
	checkAllowances(300, 0, 0, 0, 0, 0)

	urequire.Error(t, ledger.SpendAllowance(alice, bob, 2000000))
	checkBalances(1000, 0, 100)
	checkAllowances(300, 0, 0, 0, 0, 0)

	urequire.NoError(t, ledger.SpendAllowance(alice, bob, 100))
	checkBalances(1000, 0, 100)
	checkAllowances(200, 0, 0, 0, 0, 0)
}

func TestCallerTeller(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carl := testutils.TestAddress("carl")

	token, ledger := newTestToken("Dummy", "DUMMY", 6, 0, cur)
	teller := ledger.CallerTeller()

	checkBalances := func(aliceEB, bobEB, carlEB uint256) {
		t.Helper()
		exp := fmtBalances(aliceEB, bobEB, carlEB)
		got := fmtBalances(token.BalanceOf(alice), token.BalanceOf(bob), token.BalanceOf(carl))
		uassert.Equal(t, got, exp, "invalid balances")
	}
	checkAllowances := func(abEB, acEB, baEB, bcEB, caEB, cbEB uint256) {
		t.Helper()
		exp := fmtAllowances(abEB, acEB, baEB, bcEB, caEB, cbEB)
		got := fmtAllowances(
			token.Allowance(alice, bob), token.Allowance(alice, carl),
			token.Allowance(bob, alice), token.Allowance(bob, carl),
			token.Allowance(carl, alice), token.Allowance(carl, bob),
		)
		uassert.Equal(t, got, exp, "invalid allowances")
	}

	urequire.NoError(t, ledger.Mint(alice, 1000))
	checkBalances(1000, 0, 0)
	checkAllowances(0, 0, 0, 0, 0, 0)

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) { urequire.NoError(t, teller.Approve(0, cur, bob, 600)) }(cross(cur))
	checkBalances(1000, 0, 0)
	checkAllowances(600, 0, 0, 0, 0, 0)

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) { urequire.Error(t, teller.TransferFrom(0, cur, alice, carl, 700)) }(cross(cur))
	checkBalances(1000, 0, 0)
	checkAllowances(600, 0, 0, 0, 0, 0)
	func(cur realm) { urequire.NoError(t, teller.TransferFrom(0, cur, alice, carl, 400)) }(cross(cur))
	checkBalances(600, 0, 400)
	checkAllowances(200, 0, 0, 0, 0, 0)
}

func TestReadonlyTeller(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	token, ledger := newTestToken("Dummy", "DUMMY", 6, 0, cur)
	urequire.NoError(t, ledger.Mint(alice, math.MaxUint256))
	ro := token.ReadonlyTeller()

	uassert.True(t, ro.TotalSupply() == math.MaxUint256, "readonly TotalSupply")
	uassert.True(t, ro.BalanceOf(alice) == math.MaxUint256, "readonly BalanceOf")

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		uassert.ErrorIs(t, ro.Transfer(0, cur, bob, 1), ErrReadonly)
		uassert.ErrorIs(t, ro.Approve(0, cur, bob, 1), ErrReadonly)
		uassert.ErrorIs(t, ro.TransferFrom(0, cur, alice, bob, 1), ErrReadonly)
	}(cross(cur))
	uassert.True(t, token.BalanceOf(bob) == 0, "readonly teller moved funds")
}

// TestCallerTellerAbove2Pow63 drives the frame-relative teller with amounts
// beyond the int64 range, including an approval of MaxUint256.
func TestCallerTellerAbove2Pow63(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carl := testutils.TestAddress("carl")

	token, ledger := newTestToken("Dummy", "DUMMY", 18, 0, cur)
	teller := ledger.CallerTeller()

	big := uint256(1) << 200
	urequire.NoError(t, ledger.Mint(alice, big))

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		urequire.NoError(t, teller.Transfer(0, cur, bob, big/2))
		urequire.NoError(t, teller.Approve(0, cur, bob, math.MaxUint256))
	}(cross(cur))
	checkEq := func(want, got uint256, msg string) {
		t.Helper()
		if want != got {
			t.Fatalf("%s: want %s, got %s", msg, strconv.FormatUint256(want, 10), strconv.FormatUint256(got, 10))
		}
	}
	checkEq(big/2, token.BalanceOf(alice), "alice")
	checkEq(big/2, token.BalanceOf(bob), "bob")
	checkEq(math.MaxUint256, token.Allowance(alice, bob), "allowance")

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		// One more than alice holds is refused.
		uassert.ErrorIs(t, teller.TransferFrom(0, cur, alice, carl, big/2+1), ErrInsufficientBalance)
		urequire.NoError(t, teller.TransferFrom(0, cur, alice, carl, big/2))
	}(cross(cur))
	checkEq(0, token.BalanceOf(alice), "alice drained")
	checkEq(big/2, token.BalanceOf(carl), "carl")
	checkEq(math.MaxUint256-big/2, token.Allowance(alice, bob), "allowance decreased")
}
