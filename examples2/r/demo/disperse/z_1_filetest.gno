// PKGPATH: gno.land/r/demo/main

// SEND: 300ugnot

package main

import (
	"std"

	"gno.land/r/demo/disperse"
)

func main() {
	mainAddr := std.DerivePkgAddr("gno.land/r/demo/main")
	disperseAddr := std.DerivePkgAddr("gno.land/r/demo/disperse")
	beneficiary1 := std.Address("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0")
	beneficiary2 := std.Address("g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c")

	banker := std.NewBanker(std.BankerTypeRealmSend)
	println("main balance before send:", banker.GetCoins(mainAddr))
	println("disperse balance before send:", banker.GetCoins(disperseAddr))

	banker.SendCoins(mainAddr, disperseAddr, std.Coins{{"ugnot", 200}})
	println("main balance after send:", banker.GetCoins(mainAddr))
	println("disperse balance after send:", banker.GetCoins(disperseAddr))

	addressesStr := beneficiary1.String() + "," + beneficiary2.String()
	disperse.DisperseUgnotString(cross, addressesStr, "150,50")

	println("main balance after disperse:", banker.GetCoins(mainAddr))
	println("disperse balance after disperse:", banker.GetCoins(disperseAddr))
	println("beneficiary1 balance:", banker.GetCoins(beneficiary1))
	println("beneficiary2 balance:", banker.GetCoins(beneficiary2))
}

// Output:
// main balance before send: 300ugnot
// disperse balance before send:
// main balance after send: 100ugnot
// disperse balance after send: 200ugnot
// main balance after disperse: 100ugnot
// disperse balance after disperse:
// beneficiary1 balance: 150ugnot
// beneficiary2 balance: 50ugnot
