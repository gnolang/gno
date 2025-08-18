// PKGPATH: gno.land/r/demo/bank1

package banktest

import (
	"std"
	"testing"
)

func main() {
	banktestAddr := std.DerivePkgAddr("gno.land/r/archives/banktest")

	mainaddr := std.DerivePkgAddr("gno.land/r/demo/bank1")
	testing.SetOriginCaller(mainaddr)

	banker := std.NewBanker(std.BankerTypeRealmSend)
	send := std.Coins{{"ugnot", 123}}
	banker.SendCoins(banktestAddr, mainaddr, send)
}

// Error:
// expected package name [bank1] but got [banktest]
