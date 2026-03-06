// PKGPATH: gno.land/r/demo/bank1

package bank1

import (
	"chain"
	"chain/banker"
	"testing"

	"gno.land/r/archive/banktest"
)

func main() {
	banktestAddr := chain.PackageAddress("gno.land/r/archive/banktest")
	mainaddr := chain.PackageAddress("gno.land/r/demo/bank1")

	// print main balance before.
	testing.SetOriginCaller(mainaddr)

	banker_ := banker.NewBanker(banker.BankerTypeReadonly)
	mainbal := banker_.GetCoins(mainaddr)
	println("main before:", mainbal) // plus OriginSend equals 300.

	// simulate a Deposit call.
	testing.IssueCoins(banktestAddr, chain.Coins{{"ugnot", 100000000}})
	testing.SetOriginSend(chain.Coins{{"ugnot", 100000000}})
	testing.SetRealm(testing.NewUserRealm(mainaddr))
	res := banktest.Deposit(cross, "ugnot", 55000000)
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
