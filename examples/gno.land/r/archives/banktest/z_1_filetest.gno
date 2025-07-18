// PKGPATH: gno.land/r/demo/bank1

package banktest

import (
	"std"
	"testing"

	"gno.land/r/demo/banktest"
)

func main() {
	banktestAddr := std.DerivePkgAddr("gno.land/r/demo/banktest")
	mainaddr := std.DerivePkgAddr("gno.land/r/demo/bank1")

	// simulate a Deposit call.
	testing.IssueCoins(banktestAddr, std.Coins{{"ugnot", 100000000}})
	testing.SetOriginSend(std.Coins{{"ugnot", 100000000}})
	testing.SetRealm(std.NewUserRealm(mainaddr))

	res := banktest.Deposit(cross, "ugnot", 101000000)
	println(res)
}

// Error:
// cannot send "101000000ugnot", limit "100000000ugnot" exceeded with "" already spent
