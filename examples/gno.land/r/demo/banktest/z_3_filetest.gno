package main

import (
	"std"
)

func main() {
	banktestAddr := std.DerivePkgAddr("gno.land/r/demo/banktest")

	// print main balance before.
	mainaddr := std.DerivePkgAddr("main")
	std.TestSetOrigCaller(mainaddr)

	banker := std.GetBanker(std.BankerTypeRealmSend)
	send := std.Coins{{"ugnot", int64(123)}}
	banker.SendCoins(banktestAddr, mainaddr, send)

}

// Error:
// can only send coins from realm that created banker "g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4", not "g1dv3435088tlrgggf745kaud0ptrkc9v42k8llz"
