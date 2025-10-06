// PKGPATH: gno.land/r/aib/bankertest
// SEND: 100ugnot
package bankertest

import (
	"chain"
	"chain/banker"
	"chain/runtime"

	"gno.land/p/aib/limitedbanker"
	"gno.land/p/nt/testutils"
)

// limit to 0
func main() {
	var (
		banker = limitedbanker.New(
			banker.NewBanker(banker.BankerTypeRealmSend),
			nil,
		)
		from = runtime.CurrentRealm().Address()
		to   = testutils.TestAddress("test")
	)
	println("balance:", banker.GetCoins(from))
	println("spendable:", banker.SpendableCoins())

	banker.SendCoins(from, to, chain.NewCoins(chain.NewCoin("ugnot", 1)))
}

// Output:
// balance: 100ugnot
// spendable: 0ugnot

// Error:
// spendLimit 0ugnot exceeded for coins 1ugnot
