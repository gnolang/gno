// PKGPATH: gno.land/r/demo/main

// SEND: 300ugnot

package main

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"testing"

	"gno.land/r/demo/disperse"
)

func main() {
	disperseAddr := chain.DerivePkgAddr("gno.land/r/demo/disperse")
	mainaddr := chain.DerivePkgAddr("gno.land/r/demo/main")

	testing.SetOriginCaller(mainaddr)

	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)

	banker_.SendCoins(mainaddr, disperseAddr, banker.Coins{{"ugnot", 100}})
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/demo/disperse"))
	disperse.DisperseUgnotString("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0,g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c", "150,50")
}

// Error:
// disperse: mismatch between coins sent and params called
