// PKGPATH: gno.land/r/demo/grc20kvseam

// Runs the full token flow over each pluggable KV backend — the avl default,
// hashmap, and bptree — verifying that the WithStorage seam is
// backend-agnostic end to end.
//
// This is a filetest rather than a grc20 _test.gno file on purpose:
// MsgAddPackage type-checks a package's test files too, so a test-only import
// of hashmap/bptree from inside the grc20 package would pull both into the
// on-chain package closure of every realm that imports grc20 — a deploy-time
// cost paid even by default avl tokens that use neither backend. filetests/
// has no gnomod.toml, so it is never deployed and its imports stay out of
// grc20's closure.

package grc20kvseam

import (
	"strconv"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/hashmap/v0"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/testutils/v0"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
	carl  = testutils.TestAddress("carl")

	nextTokenID seqid.ID
)

func check(name, what string, got, want int64) {
	if got != want {
		panic(name + ": " + what + ": got " + strconv.FormatInt(got, 10) +
			", want " + strconv.FormatInt(want, 10))
	}
}

func runTokenFlow(cur realm, name string, opts ...grc20.Option) {
	tok, ledger := grc20.NewToken(name, "DUMMY", 4, nextTokenID.Next(), cur, opts...)

	checkBalances := func(a, b, c int64) {
		check(name, "alice balance", tok.BalanceOf(alice), a)
		check(name, "bob balance", tok.BalanceOf(bob), b)
		check(name, "carl balance", tok.BalanceOf(carl), c)
		check(name, "total supply", tok.TotalSupply(), a+b+c)
	}

	if err := ledger.Mint(alice, 1000); err != nil {
		panic(name + ": mint: " + err.Error())
	}
	checkBalances(1000, 0, 0)
	check(name, "known accounts after mint", int64(tok.KnownAccounts()), 1)

	if err := ledger.Transfer(alice, bob, 400); err != nil {
		panic(name + ": transfer: " + err.Error())
	}
	checkBalances(600, 400, 0)
	check(name, "known accounts after transfer", int64(tok.KnownAccounts()), 2)

	if err := ledger.Approve(alice, bob, 250); err != nil {
		panic(name + ": approve: " + err.Error())
	}
	check(name, "allowance", tok.Allowance(alice, bob), 250)

	if err := ledger.TransferFrom(alice, bob, carl, 100); err != nil {
		panic(name + ": transferFrom: " + err.Error())
	}
	checkBalances(500, 400, 100)
	check(name, "allowance after transferFrom", tok.Allowance(alice, bob), 150)

	if err := ledger.Burn(bob, 400); err != nil {
		panic(name + ": burn: " + err.Error())
	}
	checkBalances(500, 0, 100)
	// bob's zero balance is removed from the ledger.
	check(name, "known accounts after burn", int64(tok.KnownAccounts()), 2)

	if err := ledger.Transfer(alice, carl, 501); err == nil {
		panic(name + ": expected insufficient balance error")
	}
	if err := ledger.TransferFrom(alice, bob, carl, 151); err == nil {
		panic(name + ": expected insufficient allowance error")
	}

	println(name, "ok:", tok.KnownAccounts(), "accounts,", tok.TotalSupply(),
		"supply,", tok.Allowance(alice, bob), "allowance")
}

func main(cur realm) {
	runTokenFlow(cur, "avl-default")
	runTokenFlow(cur, "hashmap",
		grc20.WithStorage(func() grc20.KV { return hashmap.New() }))
	runTokenFlow(cur, "bptree",
		grc20.WithStorage(func() grc20.KV { return bptree.NewBPTree32() }))
}

// Output:
// avl-default ok: 2 accounts, 600 supply, 150 allowance
// hashmap ok: 2 accounts, 600 supply, 150 allowance
// bptree ok: 2 accounts, 600 supply, 150 allowance
