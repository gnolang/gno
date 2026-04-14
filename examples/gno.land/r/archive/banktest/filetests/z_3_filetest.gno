// PKGPATH: gno.land/r/demo/bank1

package bank1

import (
	"chain"
	"chain/banker"
	"testing"
)

func main() {
	banktestAddr := chain.PackageAddress("gno.land/r/archive/banktest")

	mainaddr := chain.PackageAddress("gno.land/r/demo/bank1")
	testing.SetOriginCaller(mainaddr)

	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)
	send := chain.Coins{{"ugnot", 123}}
	banker_.SendCoins(banktestAddr, mainaddr, send)
}

// Error:
// can only send coins from realm that created banker "g1tnpdmvrmtgql8fmxgsq9rwtst5hsxahk3f05dk", not "g1hsf4l9pxp4vklr6udmr29ztgqvhswjl8mt78fh"
