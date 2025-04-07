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
// cannot send "101000000ugnot", limit "100000000ugnot" exceeded with "" already spent
