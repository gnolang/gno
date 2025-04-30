// PKGPATH: gno.land/r/demo/main

// SEND: 100ugnot

package main

import (
	"std"

	"gno.land/r/demo/disperse"
)

func main() {
	mainAddr := std.DerivePkgAddr("gno.land/r/demo/main")
	beneficiary1 := std.Address("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0")
	beneficiary2 := std.Address("g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c")

	banker := std.NewBanker(std.BankerTypeRealmSend)
	println("main balance before disperse:", banker.GetCoins(mainAddr))

	addressesStr := beneficiary1.String() + "," + beneficiary2.String()
	disperse.DisperseUgnotString(addressesStr, "150,50")
}

// Error:
// disperse: mismatch between coins sent and params called
