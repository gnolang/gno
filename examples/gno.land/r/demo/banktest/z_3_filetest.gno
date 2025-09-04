// PKGPATH: gno.land/r/demo/bank1

package bank1

import (
	"chain"
	"chain/banker"
	"testing"
)

func main() {
	banktestAddr := chain.DerivePkgAddr("gno.land/r/demo/banktest")

	mainaddr := chain.DerivePkgAddr("gno.land/r/demo/bank1")
	testing.SetOriginCaller(mainaddr)

	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)
	send := banker.Coins{{"ugnot", 123}}
	banker_.SendCoins(banktestAddr, mainaddr, send)
}

// Error:
// can only send coins from realm that created banker "g1tnpdmvrmtgql8fmxgsq9rwtst5hsxahk3f05dk", not "g1dv3435088tlrgggf745kaud0ptrkc9v42k8llz"
