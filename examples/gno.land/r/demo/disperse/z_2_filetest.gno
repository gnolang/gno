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

	banker.SendCoins(mainaddr, disperseAddr, std.Coins{{"ugnot", 100}})
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/disperse"))
	disperse.DisperseUgnotString("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0,g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c", "150,50")
}

// Error:
// banker with type BankerTypeOriginSend can only be instantiated by the origin package
