// SEND: 200ugnot

package main

import (
	"std"

	"gno.land/r/demo/disperse"
)

func main() {
	disperseAddr := std.DerivePkgAddr("gno.land/r/demo/disperse")
	mainaddr := std.DerivePkgAddr("main")

	std.TestSetOrigPkgAddr(disperseAddr)
	std.TestSetOrigCaller(mainaddr)

	banker := std.GetBanker(std.BankerTypeRealmSend)

	mainbal := banker.GetCoins(mainaddr)
	println("main before:", mainbal)

	banker.SendCoins(mainaddr, disperseAddr, std.Coins{{"ugnot", 200}})
	disperse.DisperseUgnotString("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0,g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c", "150,50")

	mainbal = banker.GetCoins(mainaddr)
	println("main after:", mainbal)
}

// Output:
// main before: 200000200ugnot
// main after: 200000000ugnot
