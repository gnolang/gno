// SEND: 300ugnot

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

	banker.SendCoins(mainaddr, disperseAddr, std.Coins{{"ugnot", 100}})
	disperse.DisperseUgnotString("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0,g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c", "150,50")
}

// Error:
// balance needs to be equal to amount sent
