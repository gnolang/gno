// PKGPATH: gno.land/r/demo/main

// SEND: 100ugnot

package main

import (
	"chain"
	"chain/banker"

	"gno.land/r/demo/disperse"
)

func main() {
	mainAddr := chain.PackageAddress("gno.land/r/demo/main")
	disperseAddr := chain.PackageAddress("gno.land/r/demo/disperse")
	beneficiary1 := address("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0")
	beneficiary2 := address("g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c")

	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)
	println("main balance before send:", banker_.GetCoins(mainAddr))
	println("disperse balance before send:", banker_.GetCoins(disperseAddr))

	banker_.SendCoins(mainAddr, disperseAddr, chain.Coins{{"ugnot", 100}})
	println("main balance after send:", banker_.GetCoins(mainAddr))
	println("disperse balance after send:", banker_.GetCoins(disperseAddr))

	addressesStr := beneficiary1.String() + "," + beneficiary2.String()
	disperse.DisperseUgnotString(cross, addressesStr, "150,50")
}

// Error:
// disperse: mismatch between coins sent and params called
