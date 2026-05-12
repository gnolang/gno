// PKGPATH: gno.land/r/demo/main

package main

import (
	"chain"
	"chain/banker"
	"testing"

	"gno.land/r/demo/disperse"
)

func main() {
	mainAddr := chain.PackageAddress("gno.land/r/demo/main")
	disperseAddr := chain.PackageAddress("gno.land/r/demo/disperse")
	beneficiary1 := address("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0")
	beneficiary2 := address("g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c")

	testing.IssueCoins(mainAddr, chain.Coins{{Denom: "ugnot", Amount: 250}})
	testing.SetOriginSend(chain.Coins{{Denom: "ugnot", Amount: 250}})
	testing.SetRealm(testing.NewUserRealm(mainAddr))

	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)
	println("main balance before send:", banker_.GetCoins(mainAddr))
	println("disperse balance before send:", banker_.GetCoins(disperseAddr))

	banker_.SendCoins(mainAddr, disperseAddr, chain.Coins{{"ugnot", 250}})
	println("main balance after send:", banker_.GetCoins(mainAddr))
	println("disperse balance after send:", banker_.GetCoins(disperseAddr))

	addressesStr := beneficiary1.String() + "," + beneficiary2.String()
	disperse.DisperseUgnotString(cross, addressesStr, "150,50")

	println("main balance after disperse:", banker_.GetCoins(mainAddr))
	println("disperse balance after disperse:", banker_.GetCoins(disperseAddr))
	println("beneficiary1 balance:", banker_.GetCoins(beneficiary1))
	println("beneficiary2 balance:", banker_.GetCoins(beneficiary2))
}

// Output:
// main balance before send: 250ugnot
// disperse balance before send:
// main balance after send:
// disperse balance after send: 250ugnot
// main balance after disperse: 50ugnot
// disperse balance after disperse:
// beneficiary1 balance: 150ugnot
// beneficiary2 balance: 50ugnot
