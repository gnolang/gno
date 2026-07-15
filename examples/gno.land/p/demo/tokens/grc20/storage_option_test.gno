package grc20

import (
	"testing"

	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/hashmap/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/urequire/v0"
)

// TestWithStorageBackends runs the full token flow over each pluggable KV
// backend — the avl default, hashmap, and bptree — verifying that the
// WithStorage seam is backend-agnostic end to end.
func TestWithStorageBackends(cur realm, t *testing.T) {
	runTokenFlow(cur, t, "avl-default", nil)
	runTokenFlow(cur, t, "hashmap",
		[]Option{WithStorage(func() KV { return hashmap.New() })})
	runTokenFlow(cur, t, "bptree",
		[]Option{WithStorage(func() KV { return bptree.NewBPTree32() })})
}

func runTokenFlow(cur realm, t *testing.T, name string, opts []Option) {
	var (
		alice = testutils.TestAddress("alice")
		bob   = testutils.TestAddress("bob")
		carl  = testutils.TestAddress("carl")

		bank   *Token
		ledger *PrivateLedger
	)

	// Same-realm cross promotes the test's EOA-origin cur into a fresh
	// CodeRealm cur so NewToken's IsCurrent check passes (see newTestToken).
	func(cur realm) {
		bank, ledger = NewToken(0, cur, "Dummy", "DUMMY", 4, opts...)
	}(cross(cur))

	checkBalances := func(alice_, bob_, carl_ int64) {
		t.Helper()
		urequire.Equal(t, alice_, bank.BalanceOf(alice), name+": alice balance")
		urequire.Equal(t, bob_, bank.BalanceOf(bob), name+": bob balance")
		urequire.Equal(t, carl_, bank.BalanceOf(carl), name+": carl balance")
		urequire.Equal(t, alice_+bob_+carl_, bank.TotalSupply(), name+": total supply")
	}

	urequire.NoError(t, ledger.Mint(alice, 1000), name)
	checkBalances(1000, 0, 0)
	urequire.Equal(t, 1, bank.KnownAccounts(), name+": known accounts after mint")

	urequire.NoError(t, ledger.Transfer(alice, bob, 400), name)
	checkBalances(600, 400, 0)
	urequire.Equal(t, 2, bank.KnownAccounts(), name+": known accounts after transfer")

	urequire.NoError(t, ledger.Approve(alice, bob, 250), name)
	urequire.Equal(t, int64(250), bank.Allowance(alice, bob), name+": allowance")

	urequire.NoError(t, ledger.TransferFrom(alice, bob, carl, 100), name)
	checkBalances(500, 400, 100)
	urequire.Equal(t, int64(150), bank.Allowance(alice, bob), name+": allowance after transferFrom")

	urequire.NoError(t, ledger.Burn(bob, 400), name)
	checkBalances(500, 0, 100)
	// bob's zero balance is removed from the ledger.
	urequire.Equal(t, 2, bank.KnownAccounts(), name+": known accounts after burn")

	urequire.Error(t, ledger.Transfer(alice, carl, 501), name+": insufficient balance")
	urequire.Error(t, ledger.TransferFrom(alice, bob, carl, 151), name+": insufficient allowance")
}
