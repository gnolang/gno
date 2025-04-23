// Empty line between the directives is important for them to be parsed
// independently. :facepalm:

// PKGPATH: gno.land/r/demo/bank1

package bank1

import (
	"std"
	"testing"
)

func main() {
	banktestAddr := std.DerivePkgAddr("gno.land/r/demo/banktest")

	mainaddr := std.DerivePkgAddr("gno.land/r/demo/bank1")
	testing.SetOriginCaller(mainaddr)

	banker := std.NewBanker(std.BankerTypeRealmSend)
	send := std.Coins{{"ugnot", 123}}
	banker.SendCoins(banktestAddr, mainaddr, send)
}

// Error:
// missing crossing() after cross call in main from gno.land/r/demo/bank1 to gno.land/r/demo/bank1
