// PKGPATH: gno.land/r/demo/main

// SEND: 200ugnot

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

	mainbal := banker_.GetCoins(mainaddr)
	println("main before:", mainbal)

	banker_.SendCoins(mainaddr, disperseAddr, banker.Coins{{"ugnot", 200}})
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/demo/disperse"))
	disperse.DisperseUgnotString("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0,g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c", "150,50")

	mainbal = banker_.GetCoins(mainaddr)
	println("main after:", mainbal)
}

// Output:
// main before: 200ugnot
// main after:
