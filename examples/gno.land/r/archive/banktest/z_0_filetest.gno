// PKGPATH: gno.land/r/demo/bank1
// SEND: 100000000ugnot

package bank1

import (
	"chain"
	"chain/banker"
	"testing"

	"gno.land/r/archive/banktest"
)

func main() {
	// set up main address and banktest addr.
	banktestAddr := chain.PackageAddress("gno.land/r/archive/banktest")
	mainaddr := chain.PackageAddress("gno.land/r/demo/bank1")
	testing.SetOriginCaller(mainaddr)

	// get and print balance of mainaddr.
	// with the SEND, + 200 gnot given by the TestContext, main should have 300gnot.
	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)
	mainbal := banker_.GetCoins(mainaddr)
	println("main before:", mainbal)

	// simulate a Deposit call. use Send + OriginSend to simulate -send.
	banker_.SendCoins(mainaddr, banktestAddr, chain.Coins{{"ugnot", 100_000_000}})
	testing.SetOriginSend(chain.Coins{{"ugnot", 100_000_000}})
	testing.SetRealm(testing.NewUserRealm(mainaddr))
	res := banktest.Deposit(cross, "ugnot", 50_000_000) // bank1 can't send? should be r/demo/bank1 to r/demo/banktest, is bank1 -> bank1.
	println("Deposit():", res)

	// print main balance after.
	mainbal = banker_.GetCoins(mainaddr)
	println("main after:", mainbal)

	// simulate a Render(). banker should have given back all coins.
	res = banktest.Render("")
	println(res)
}

// Output:
// main before: 100000000ugnot
// Deposit(): returned!
// main after: 50000000ugnot
// ## recent activity
//
//  * g1tnpdmvrmtgql8fmxgsq9rwtst5hsxahk3f05dk 100000000ugnot sent, 50000000ugnot returned, at 2009-02-13 11:31pm UTC
//
// ## total deposits
// 50000000ugnot
