// SEND: 100000000ugnot

package main

import (
	"std"

	"gno.land/r/demo/banktest"
)

func main() {
	// set up main address and banktest addr.
	banktestAddr := std.DerivePkgAddr("gno.land/r/demo/banktest")
	mainaddr := std.DerivePkgAddr("main")
	std.TestSetOrigCaller(mainaddr)
	std.TestSetOrigPkgAddr(banktestAddr)

	// get and print balance of mainaddr.
	// with the SEND, + 200 gnot given by the TestContext, main should have 300gnot.
	banker := std.GetBanker(std.BankerTypeRealmSend)
	mainbal := banker.GetCoins(mainaddr)
	println("main before:", mainbal)

	// simulate a Deposit call. use Send + OrigSend to simulate -send.
	banker.SendCoins(mainaddr, banktestAddr, std.Coins{{"ugnot", 100_000_000}})
	std.TestSetOrigSend(std.Coins{{"ugnot", 100_000_000}}, nil)
	res := banktest.Deposit("ugnot", 50_000_000)
	println("Deposit():", res)

	// print main balance after.
	mainbal = banker.GetCoins(mainaddr)
	println("main after:", mainbal)

	// simulate a Render(). banker should have given back all coins.
	res = banktest.Render("")
	println(res)
}

// Output:
// main before: 300000000ugnot
// Deposit(): returned!
// main after: 250000000ugnot
// ## recent activity
//
//  * g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4 100000000ugnot sent, 50000000ugnot returned, at 2009-02-13 11:31pm UTC
//
// ## total deposits
// 50000000ugnot
