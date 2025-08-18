// PKGPATH: gno.land/r/demo/bank1

package banktest

import (
	"std"
	"testing"

	"gno.land/r/archives/banktest"
)

func main() {
	banktestAddr := std.DerivePkgAddr("gno.land/r/archives/banktest")
	mainaddr := std.DerivePkgAddr("gno.land/r/demo/bank1")

	// print main balance before.
	testing.SetOriginCaller(mainaddr)

	banker := std.NewBanker(std.BankerTypeReadonly)
	mainbal := banker.GetCoins(mainaddr)
	println("main before:", mainbal) // plus OriginSend equals 300.

	// simulate a Deposit call.
	testing.IssueCoins(banktestAddr, std.Coins{{"ugnot", 100000000}})
	testing.SetOriginSend(std.Coins{{"ugnot", 100000000}})
	testing.SetRealm(std.NewUserRealm(mainaddr))
	res := banktest.Deposit(cross, "ugnot", 55000000)
	println("Deposit():", res)

	// print main balance after.
	mainbal = banker.GetCoins(mainaddr)
	println("main after:", mainbal) // now 255.

	// simulate a Render().
	res = banktest.Render("")
	println(res)
}

// Error:
// expected package name [bank1] but got [banktest]
