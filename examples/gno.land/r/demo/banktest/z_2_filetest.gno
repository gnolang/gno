// PKGPATH: gno.land/r/demo/bank1
// SEND: 200000000ugnot

package bank1

import (
	"std"

	"gno.land/r/demo/banktest"
)

func main() {
	banktestAddr := std.DerivePkgAddr("gno.land/r/demo/banktest")

	// print main balance before.
	mainaddr := std.DerivePkgAddr("gno.land/r/demo/bank1")
	std.TestSetOrigCaller(mainaddr)

	banker := std.GetBanker(std.BankerTypeReadonly)
	mainbal := banker.GetCoins(mainaddr)
	println("main before:", mainbal)

	// simulate a Deposit call.
	std.TestSetOrigPkgAddr(banktestAddr)
	std.TestIssueCoins(banktestAddr, std.Coins{{"ugnot", 100000000}})
	std.TestSetOrigSend(std.Coins{{"ugnot", 100000000}}, nil)
	res := banktest.Deposit("ugnot", 55000000)
	println("Deposit():", res)

	// print main balance after.
	mainbal = banker.GetCoins(mainaddr)
	println("main after:", mainbal) // now 255.

	// simulate a Render().
	res = banktest.Render("")
	println(res)
}

// Output:
// main before: 200000000ugnot
// Deposit(): returned!
// main after: 255000000ugnot
// ## recent activity
//
//  * g1tnpdmvrmtgql8fmxgsq9rwtst5hsxahk3f05dk 100000000ugnot sent, 55000000ugnot returned, at 2021-01-29 5:00pm UTC
//
// ## total deposits
// 45000000ugnot
