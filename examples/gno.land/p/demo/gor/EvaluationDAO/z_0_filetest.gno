package treasury

// NOTE: this doesn't do anything, as it sends to "main".
// SEND: 100000000ugnot

import (
	"gno.land/p/demo/treasury"
	"std"
)

func main() {
	treasuryAddr := std.DerivePkgAddr("gno.land/p/demo/gor")

	// print main balance before.
	mainaddr := std.DerivePkgAddr("main")
	std.TestSetOrigCaller(mainaddr)

	banker := std.GetBanker(std.BankerTypeReadonly)
	mainbal := banker.GetCoins(mainaddr)
	println("main before:", mainbal) // plus OrigSend equals 300.

	// simulate a Deposit call.
	std.TestSetOrigPkgAddr(treasuryAddr)
	std.TestIssueCoins(banktestAddr, std.Coins{{"ugnot", 100000000}})
	std.TestSetOrigSend(std.Coins{{"ugnot", 100000000}}, nil)
	res := treasury.CollectFee("ugnot", 100000000)
	println("Collect():", res)

	// print main balance after.
	//mainbal = banker.GetCoins(mainaddr)
	//println("main after:", mainbal) // still 300.

	// simulate a Render().
	//res = banktest.Render("")
	//println(res)
}

// Output:
// main before: 200000000ugnot
// Deposit(): returned!
// main after: 300000000ugnot
// ## recent activity
//
//  * g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4 100000000ugnot sent, 100000000ugnot returned, at 2009-02-13 11:31pm UTC
//
// ## total deposits
