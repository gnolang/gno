// PKGPATH: gno.land/r/demo/bank1

package bank1

import (
	"chain"
	"testing"

	"gno.land/r/archive/banktest"
)

func main() {
	banktestAddr := chain.PackageAddress("gno.land/r/archive/banktest")
	mainaddr := chain.PackageAddress("gno.land/r/demo/bank1")

	// simulate a Deposit call.
	testing.IssueCoins(banktestAddr, chain.Coins{{"ugnot", 100000000}})
	testing.SetOriginSend(chain.Coins{{"ugnot", 100000000}})
	testing.SetRealm(testing.NewUserRealm(mainaddr))

	res := banktest.Deposit(cross, "ugnot", 101000000)
	println(res)
}

// Error:
// cannot send "101000000ugnot", limit "100000000ugnot" exceeded with "" already spent
