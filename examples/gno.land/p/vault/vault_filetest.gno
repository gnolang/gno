package main

import (
	"std"
	"time"

	"gno.land/p/grc/grc20"
	"gno.land/p/testutils"
	"gno.land/p/ufmt"
	"gno.land/p/vault"
)

func main() {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob") // recovery address.
	pkgaddr := std.GetOrigPkgAddr()
	std.TestSetOrigCaller(alice)

	fooAdminToken := grc20.NewAdminToken("Foo", "FOO", 4)
	fooAdminToken.Mint(alice, 1000)
	fooToken := fooAdminToken.GRC20()

	printBalances := func() {
		aliceBalance, _ := fooToken.BalanceOf(alice)
		bobBalance, _ := fooToken.BalanceOf(bob)
		pkgBalance, _ := fooToken.BalanceOf(pkgaddr)
		println(ufmt.Sprintf(
			"balances: alice=%d, bob=%d, pkg=%d",
			aliceBalance, bobBalance, pkgBalance,
		))
	}

	lockAmount := uint(300)
	lockDuration := uint(5)
	printBalances()

	v := vault.New(fooToken, lockAmount, bob, lockDuration)
	checkErr(v.Unvault(200))
	printBalances()

	checkErr(v.Redeem())
	printBalances()

	println(v, alice, bob)
	// TODO: example of standard flow.
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Output:
// balances: alice=1000, bob=0, pkg=0
// from g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh to g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4 amount 300 fromBalance 1000 toBalance 0
// balances: alice=700, bob=0, pkg=300
// balances: alice=700, bob=0, pkg=300
// &0xc000905500.(*gno.land/p/vault.impl) g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh g1vfhkyh6lta047h6lta047h6lta047h6l03vdhu
