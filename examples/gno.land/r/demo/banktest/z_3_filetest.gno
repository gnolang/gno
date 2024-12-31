// Empty line between the directives is important for them to be parsed
// independently. :facepalm:

// PKGPATH: gno.land/r/demo/bank1

package bank1

import (
	"std"
)

func main() {
	banktestAddr := std.DerivePkgAddr("gno.land/r/demo/banktest")

	mainaddr := std.DerivePkgAddr("gno.land/r/demo/bank1")
	std.TestSetOrigCaller(mainaddr)

	banker := std.GetBanker(std.BankerTypeRealmSend)
	send := std.Coins{{"ugnot", 123}}
	banker.SendCoins(banktestAddr, mainaddr, send)
}

// Error:
// can only send coins from realm that created banker "g1tnpdmvrmtgql8fmxgsq9rwtst5hsxahk3f05dk", not "g1dv3435088tlrgggf745kaud0ptrkc9v42k8llz"
