package main

import (
	"std"

	"gno.land/r/banktest"
)

func main() {
	banktestAddr := std.TestDerivePkgAddr("gno.land/r/banktest")

	// print main balance before.
	mainaddr := std.TestDerivePkgAddr("main")
	banker := std.GetBanker(std.BankerTypeReadonly)
	mainbal := banker.GetCoins(mainaddr)
	println("main before:", mainbal) // plus OrigSend equals 300.

	// simulate a Deposit call.
	std.TestSetOrigPkgAddr(banktestAddr)
	std.TestIssueCoins(banktestAddr, std.Coins{{"gnot", 100}})
	std.TestSetOrigSend(std.Coins{{"gnot", 100}}, nil)
	res := banktest.Deposit("gnot", 55)
	println("Deposit():", res)

	// print main balance after.
	mainbal = banker.GetCoins(mainaddr)
	println("main after:", mainbal) // now 255.

	// simulate a Render().
	res = banktest.Render("")
	println(res)
}

// Output:
// main before: 200gnot
// Deposit(): returned!
// main after: 255gnot
// ## recent activity
//
//  * g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4 100gnot sent, 55gnot returned, at 1970-01-01 12:00am UTC
//
// ## total deposits
// 45gnot
//
