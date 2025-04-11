// PKGPATH: gno.land/r/demo/main

// SEND: 300ugnot

package main

import (
	"std"
	"testing"

	"gno.land/r/demo/disperse"
)

func main() {
	disperseAddr := std.DerivePkgAddr("gno.land/r/demo/disperse")
	mainaddr := std.DerivePkgAddr("gno.land/r/demo/main")

	testing.SetOriginCaller(mainaddr)

	banker := std.NewBanker(std.BankerTypeRealmSend)

	mainbal := banker.GetCoins(mainaddr)
	println("main before:", mainbal)

	banker.SendCoins(mainaddr, disperseAddr, std.Coins{{"ugnot", 300}})
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/disperse"))
	disperse.DisperseUgnotString("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0,g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c", "150,50")

	mainbal = banker.GetCoins(mainaddr)
	println("main after:", mainbal)
}

// Output:
// main before: 300ugnot
// main after: 100ugnot
