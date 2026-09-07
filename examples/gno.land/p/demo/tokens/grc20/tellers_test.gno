package grc20

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestCallerTellerImpl(cur realm, t *testing.T) {
	tok, ledger := newTestToken("Dummy", "DUMMY", 4, 0, cur)
	teller := ledger.CallerTeller()
	urequire.False(t, tok == nil)
	var _ Teller = teller
}

// The sub-realm case (guardWrite must not lock out the token's own realm when it
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

func TestIsCanonicalUserTeller(cur realm, t *testing.T) {
	const host = "gno.land/r/demo/relay"
	token, ledger := newTestToken("Relay", "RELAY", 4, 0, cur)
	other, otherLedger := newTestToken("Relay", "RELAY", 4, 0, cur)
	urequire.Equal(t, token.ID(), other.ID())
	relay := ledger.UserTeller(host)
	uassert.True(t, IsCanonicalUserTeller(relay, token, host))
	var typedNil *fnTeller
	for _, invalid := range []Teller{
		nil, typedNil, &evilWrap{Teller: relay}, otherLedger.UserTeller(host),
		ledger.UserTeller("gno.land/r/demo/other"), ledger.CallerTeller(), token.ReadonlyTeller(),
		token.RealmTeller(0, cur), token.RealmSubTeller(0, cur, "sub"),
		ledger.ImpersonateTeller(testutils.TestAddress("alice")),
	} {
		uassert.False(t, IsCanonicalUserTeller(invalid, token, host))
	}
	uassert.False(t, IsCanonicalUserTeller(relay, nil, host))
	uassert.False(t, IsCanonicalUserTeller(relay, token, ""))
	uassert.PanicsContains(t, cur, "host cannot be empty", func() { ledger.UserTeller("") })
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
	teller := ledger.CallerTeller()

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

func TestCallerTellerRejectsForeignRealm(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	token, ledger := newTestToken("Home Guard", "HOME", 6, 0, cur)
	teller := ledger.CallerTeller()
	urequire.NoError(t, ledger.Mint(alice, 1_000))

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/foreign"))
	urequire.ErrorIs(t, teller.Transfer(0, cur, bob, 100), ErrForeignCallerTeller)
	uassert.Equal(t, int64(1_000), token.BalanceOf(alice))
	uassert.Equal(t, int64(0), token.BalanceOf(bob))
}

// EOA classification alone does not authorize a host holding a leaked teller.
func TestUserTellerHostGuard(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	_, ledger := newTestToken("Guard", "GUARD", 6, 0, cur)

	// A foreign realm whose immediate caller is the EOA alice: the empty prev
	// pkgPath makes IsUserCall true, so only the host check objects.
	eoa := testing.MakeRealm(alice, "", testing.OriginRealm())
	foreign := testing.MakeRealm(
		chain.PackageAddress("gno.land/r/demo/foreign"),
		"gno.land/r/demo/foreign",
		eoa,
	)

	teller := ledger.UserTeller("gno.land/r/demo/token").(*fnTeller)
	uassert.ErrorIs(t, teller.guardWrite(0, foreign), ErrForeignUserTeller)
	delegated := ledger.UserTeller(foreign.PkgPath()).(*fnTeller)
	uassert.NoError(t, delegated.guardWrite(0, foreign))
	for _, suffix := range []string{"evil", "/child"} {
		wrongHost := testing.MakeRealm(foreign.Address(), foreign.PkgPath()+suffix, eoa)
		uassert.ErrorIs(t, delegated.guardWrite(0, wrongHost), ErrForeignUserTeller)
	}
	sub := testing.MakeRealm(foreign.Address(), foreign.PkgPath()+"#sub", eoa)
	uassert.NoError(t, delegated.guardWrite(0, sub))
}
