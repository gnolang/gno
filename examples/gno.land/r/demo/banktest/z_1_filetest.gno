// Empty line between the directives is important for them to be parsed
// independently. :facepalm:

// PKGPATH: gno.land/r/demo/bank1

package bank1

import (
	"std"
	"testing"

	"gno.land/r/demo/banktest"
)

func main() {
	banktestAddr := std.DerivePkgAddr("gno.land/r/demo/banktest")

	// simulate a Deposit call.
	testing.IssueCoins(banktestAddr, std.Coins{{"ugnot", 100000000}})
	testing.SetOriginSend(std.Coins{{"ugnot", 100000000}})
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/banktest"))
	res := banktest.Deposit("ugnot", 101000000)
	println(res)
}

// Error:
// missing crossing() after cross call in main from gno.land/r/demo/bank1 to gno.land/r/demo/bank1
