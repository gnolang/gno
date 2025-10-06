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

// Spend some coins then exceed the limit
func main() {
	var (
		banker = limitedbanker.New(
			banker.NewBanker(banker.BankerTypeRealmSend),
			chain.NewCoins(chain.NewCoin("ugnot", 10)),
		)
		from = runtime.CurrentRealm().Address()
		to   = testutils.TestAddress("test")
	)
	println("balance:", banker.GetCoins(from))
	println("spendable:", banker.SpendableCoins())

	banker.SendCoins(from, to, chain.NewCoins(chain.NewCoin("ugnot", 1)))

	println("balance:", banker.GetCoins(from))
	println("spendable:", banker.SpendableCoins())

	banker.SendCoins(from, to, chain.NewCoins(chain.NewCoin("ugnot", 10)))
}

// Output:
// balance: 100ugnot
// spendable: 10ugnot
// balance: 99ugnot
// spendable: 9ugnot

// Error:
// spendLimit 9ugnot exceeded for coins 10ugnot
