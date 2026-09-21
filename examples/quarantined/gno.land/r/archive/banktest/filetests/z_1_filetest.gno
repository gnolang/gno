// PKGPATH: gno.land/r/demo/bank1

package bank1

import (
	"chain"
	"testing"

	"gno.land/r/archive/banktest"
)

func main(cur realm) {
	banktestAddr := chain.PackageAddress("gno.land/r/archive/banktest")
	mainaddr := chain.PackageAddress("gno.land/r/demo/bank1")

	// simulate a Deposit call.
	testing.IssueCoins(banktestAddr, chain.Coins{{"ugnot", 100000000}})
	testing.SetOriginSend(chain.Coins{{"ugnot", 100000000}})
	// The -send envelope lands on banktest, so name banktest as the payee.
	// Must come before the user-realm line below: after it, cur is a code
	// realm and banker.NewBanker(BankerTypeOriginSend, cur) is refused.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/archive/banktest"))
	testing.SetRealm(testing.NewUserRealm(mainaddr))

	res := banktest.Deposit(cross(cur), "ugnot", 101000000)
	println(res)
}

// Error:
// cannot send "101000000ugnot", limit "100000000ugnot" exceeded with "" already spent
