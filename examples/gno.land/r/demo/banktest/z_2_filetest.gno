// Empty line between the directives is important for them to be parsed
// independently. :facepalm:

// PKGPATH: gno.land/r/demo/bank1

package bank1

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"testing"

	"gno.land/r/demo/banktest"
)

func main() {
	banktestAddr := chain.DerivePkgAddr("gno.land/r/demo/banktest")

	// print main balance before.
	mainaddr := chain.DerivePkgAddr("gno.land/r/demo/bank1")
	testing.SetOriginCaller(mainaddr)

	banker_ := banker.NewBanker(banker.BankerTypeReadonly)
	mainbal := banker_.GetCoins(mainaddr)
	println("main before:", mainbal) // plus OriginSend equals 300.

	// simulate a Deposit call.
	testing.IssueCoins(banktestAddr, banker.Coins{{"ugnot", 100000000}})
	testing.SetOriginSend(banker.Coins{{"ugnot", 100000000}})
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/demo/banktest"))
	res := banktest.Deposit("ugnot", 55000000)
	println("Deposit():", res)

	// print main balance after.
	mainbal = banker_.GetCoins(mainaddr)
	println("main after:", mainbal) // now 255.

	// simulate a Render().
	res = banktest.Render("")
	println(res)
}

// Output:
// main before:
// Deposit(): returned!
// main after: 55000000ugnot
// ## recent activity
//
//  * g1tnpdmvrmtgql8fmxgsq9rwtst5hsxahk3f05dk 100000000ugnot sent, 55000000ugnot returned, at 2009-02-13 11:31pm UTC
//
// ## total deposits
// 45000000ugnot
