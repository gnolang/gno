// PKGPATH: gno.land/r/demo/bank1
// SEND: 100000000ugnot

package banktest

import (
	"std"
	"testing"

	"gno.land/r/archives/banktest"
)

func main() {
	// set up main address and banktest addr.
	banktestAddr := std.DerivePkgAddr("gno.land/r/archives/banktest")
	mainaddr := std.DerivePkgAddr("gno.land/r/demo/bank1")
	testing.SetOriginCaller(mainaddr)

	// get and print balance of mainaddr.
	// with the SEND, + 200 gnot given by the TestContext, main should have 300gnot.
	banker := std.NewBanker(std.BankerTypeRealmSend)
	mainbal := banker.GetCoins(mainaddr)
	println("main before:", mainbal)

	// simulate a Deposit call. use Send + OriginSend to simulate -send.
	banker.SendCoins(mainaddr, banktestAddr, std.Coins{{"ugnot", 100_000_000}})
	testing.SetOriginSend(std.Coins{{"ugnot", 100_000_000}})
	testing.SetRealm(std.NewUserRealm(mainaddr))
	res := banktest.Deposit(cross, "ugnot", 50_000_000) // bank1 can't send? should be r/demo/bank1 to r/demo/banktest, is bank1 -> bank1.
	println("Deposit():", res)

	// print main balance after.
	mainbal = banker.GetCoins(mainaddr)
	println("main after:", mainbal)

	// simulate a Render(). banker should have given back all coins.
	res = banktest.Render("")
	println(res)
}

// Error:
// expected package name [bank1] but got [banktest]
