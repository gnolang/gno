// PKGPATH: gno.land/r/demo/main

// SEND: 300ugnot

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

	println("main balance after disperse:", banker.GetCoins(mainAddr))
	println("beneficiary1 balance:", banker.GetCoins(beneficiary1))
	println("beneficiary2 balance:", banker.GetCoins(beneficiary2))
}

// Output:
// main balance before disperse: 300ugnot
// main balance after disperse: 100ugnot
// beneficiary1 balance: 150ugnot
// beneficiary2 balance: 50ugnot
