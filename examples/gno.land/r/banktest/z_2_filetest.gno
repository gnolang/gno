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
	std.TestIssueCoins(banktestAddr, std.Coins{{"ugnot", 100}})
	std.TestSetOrigSend(std.Coins{{"ugnot", 100}}, nil)
	res := banktest.Deposit("ugnot", 55)
	println("Deposit():", res)

	// print main balance after.
	mainbal = banker.GetCoins(mainaddr)
	println("main after:", mainbal) // now 255.

	// simulate a Render().
	res = banktest.Render("")
	println(res)
}

// Output:
// main before: 200ugnot
// Deposit(): returned!
// main after: 255ugnot
// ## recent activity
//
//  * g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4 100ugnot sent, 55ugnot returned, at 1970-01-01 12:00am UTC
//
// ## total deposits
// 45ugnot
//
