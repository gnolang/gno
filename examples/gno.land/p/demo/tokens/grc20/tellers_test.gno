package grc20

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestCallerTellerImpl(cur realm, t *testing.T) {
	tok, _ := newTestToken("Dummy", "DUMMY", 4, 0, cur)
	teller := tok.CallerTeller()
	urequire.False(t, tok == nil)
	var _ Teller = teller
}

// evilWrap embeds Teller; the seal-marker pattern (if we had one) would be
// satisfied via promotion, but IsCanonicalTeller rejects it because *evilWrap
// is nominally distinct from *fnTeller.
type evilWrap struct {
	Teller
}

func TestIsCanonicalTeller(cur realm, t *testing.T) {
	tok, _ := newTestToken("Dummy", "DUMMY", 4, 0, cur)
	legit := tok.CallerTeller()

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

	checkBalances := func(aliceEB, bobEB, carlEB int64) {
		t.Helper()
		exp := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceEB, bobEB, carlEB)
		aliceGB := token.BalanceOf(alice)
		bobGB := token.BalanceOf(bob)
		carlGB := token.BalanceOf(carl)
		got := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceGB, bobGB, carlGB)
		uassert.Equal(t, got, exp, "invalid balances")
	}
	checkAllowances := func(abEB, acEB, baEB, bcEB, caEB, cbEB int64) {
		t.Helper()
		exp := ufmt.Sprintf("ab=%d ac=%d ba=%d bc=%d ca=%d cb=%s", abEB, acEB, baEB, bcEB, caEB, cbEB)
		abGB := token.Allowance(alice, bob)
		acGB := token.Allowance(alice, carl)
		baGB := token.Allowance(bob, alice)
		bcGB := token.Allowance(bob, carl)
		caGB := token.Allowance(carl, alice)
		cbGB := token.Allowance(carl, bob)
		got := ufmt.Sprintf("ab=%d ac=%d ba=%d bc=%d ca=%d cb=%s", abGB, acGB, baGB, bcGB, caGB, cbGB)
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
	teller := token.CallerTeller()

	checkBalances := func(aliceEB, bobEB, carlEB int64) {
		t.Helper()
		exp := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceEB, bobEB, carlEB)
		aliceGB := token.BalanceOf(alice)
		bobGB := token.BalanceOf(bob)
		carlGB := token.BalanceOf(carl)
		got := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceGB, bobGB, carlGB)
		uassert.Equal(t, got, exp, "invalid balances")
	}
	checkAllowances := func(abEB, acEB, baEB, bcEB, caEB, cbEB int64) {
		t.Helper()
		exp := ufmt.Sprintf("ab=%d ac=%d ba=%d bc=%d ca=%d cb=%s", abEB, acEB, baEB, bcEB, caEB, cbEB)
		abGB := token.Allowance(alice, bob)
		acGB := token.Allowance(alice, carl)
		baGB := token.Allowance(bob, alice)
		bcGB := token.Allowance(bob, carl)
		caGB := token.Allowance(carl, alice)
		cbGB := token.Allowance(carl, bob)
		got := ufmt.Sprintf("ab=%d ac=%d ba=%d bc=%d ca=%d cb=%s", abGB, acGB, baGB, bcGB, caGB, cbGB)
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
