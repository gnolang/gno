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

	// simulate a Deposit call.
	testing.IssueCoins(banktestAddr, banker.Coins{{"ugnot", 100000000}})
	testing.SetOriginSend(banker.Coins{{"ugnot", 100000000}})
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/demo/banktest"))
	res := banktest.Deposit("ugnot", 101000000)
	println(res)
}

// Error:
// cannot send "101000000ugnot", limit "100000000ugnot" exceeded with "" already spent
