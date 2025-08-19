// PKGPATH: gno.land/r/demo/bank1

package bank1

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
// can only send coins from realm that created banker "g1tnpdmvrmtgql8fmxgsq9rwtst5hsxahk3f05dk", not "g17rhj38ydnsms35338eafnqrkgkyw7wl7nln86e"
