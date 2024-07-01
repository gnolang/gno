package main

import (
	"gno.land/p/demo/grc/exts/vault"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
	"std"
)

func main() {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")       // recovery request address (cold wallet).
	charly := testutils.TestAddress("charly") // recovery dest.
	pkgaddr := std.GetOrigPkgAddr()

	// create a fooAdminToken + fooToken (GRC20) pair.
	fooToken := grc20.NewGRC20Token("Foo", "FOO", 4)
	fooToken.Mint(alice, 1000)

	printBalances := func() {
		aliceBalance := fooToken.BalanceOf(alice)
		bobBalance := fooToken.BalanceOf(bob)
		charlyBalance := fooToken.BalanceOf(charly)
		pkgBalance := fooToken.BalanceOf(pkgaddr)
		println(ufmt.Sprintf(
			"balances: alice=%d, bob=%d, charly=%d, pkg=%d, height=%d",
			aliceBalance, bobBalance, charlyBalance, pkgBalance, std.GetHeight(),
		))
	}

	// create a vault for fooAdminToken.
	v := vault.New(fooToken)
	printBalances()

	// alice deposits 300 with an unlock duration of 5 blocks.
	std.TestSetOrigCaller(alice)
	lockAmount := uint(300)
	lockDuration := uint(5)
	checkErr(v.Deposit(lockAmount, bob, lockDuration))
	printBalances()

	// alice calls unvault for 200 tokens.
	checkErr(v.Unvault(200))
	printBalances()

	// alice waits for few blocks.
	std.TestSkipHeights(int64(lockDuration) + 1)
	printBalances()

	// alice redeems 200 tokens.
	checkErr(v.Redeem())
	printBalances()

	// bob instantly recover everything in the wallet.
	std.TestSetOrigCaller(bob)
	checkErr(v.Recover(charly))
	printBalances()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Output:
// balances: alice=1000, bob=0, charly=0, pkg=0, height=123
// balances: alice=1000, bob=0, charly=0, pkg=0, height=123
// AAA1 123 0 5
// AAA2 123 128 5
// balances: alice=1000, bob=0, charly=0, pkg=0, height=123
// balances: alice=1000, bob=0, charly=0, pkg=0, height=129
// AAA3 129 128 5
// balances: alice=1000, bob=0, charly=0, pkg=0, height=129
// balances: alice=1000, bob=0, charly=0, pkg=0, height=129
