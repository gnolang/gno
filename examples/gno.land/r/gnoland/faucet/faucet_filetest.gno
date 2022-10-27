package main

import (
	"fmt"
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/faucet"
)

func main() {
	var (
		adminaddr      = std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
		controlleraddr = testutils.TestAddress("controller")
		mainaddr       = std.TestDerivePkgAddr("main")
		faucetaddr     = std.TestDerivePkgAddr("gno.land/r/faucet")
		test1addr      = testutils.TestAddress("test1")
		test2addr      = testutils.TestAddress("test2")
		test3addr      = testutils.TestAddress("test3")
	)

	banker := std.GetBanker(std.BankerTypeReadonly)
	std.TestIssueCoins(faucetaddr, std.Coins{{"ugnot", 1000000000}})
	std.TestSetOrigPkgAddr(faucetaddr)

	showBalances := func() string {
		var (
			mainbal       = banker.GetCoins(mainaddr)
			faucetbal     = banker.GetCoins(faucetaddr)
			test1bal      = banker.GetCoins(test1addr)
			test2bal      = banker.GetCoins(test2addr)
			test3bal      = banker.GetCoins(test3addr)
			controllerbal = banker.GetCoins(controlleraddr)
			adminbal      = banker.GetCoins(adminaddr)
		)
		return fmt.Sprintf("main=%q, faucet=%q, test1=%q, test2=%q test3=%q controller=%q admin=%q",
			mainbal.String(), faucetbal.String(), test1bal.String(), test2bal.String(),
			test3bal.String(), controllerbal.String(), adminbal.String())
	}

	println("before:", showBalances())

	// simulate a Deposit call.
	std.TestSetOrigSend(std.Coins{{"ugnot", 2000000000}}, nil)
	std.TestSetOrigCaller(adminaddr)
	faucet.AdminSetControllerAddr(controlleraddr)
	std.TestSetOrigCaller(controlleraddr)
	faucet.Transfer(test1addr)
	faucet.Transfer(test2addr)
	faucet.Transfer(test1addr)

	println("after: ", showBalances())

	// simulate a Render().
	ret := faucet.Render("")
	println("---")
	println("Render: ", ret)
	println("---")
	println("done")
}

// Output:
// before: main="200000000ugnot", faucet="1000000000ugnot", test1="", test2="" test3="" controller="" admin=""
// after:  main="200000000ugnot", faucet="997000000ugnot", test1="2000000ugnot", test2="1000000ugnot" test3="" controller="" admin=""
// ---
// Render:  # Community Faucet.
//
// Status: active.
// Balance: 997000000ugnot.
// Rewarded: 3000000ugnot (in 3 times).
//
// ---
// done

// Realm:
