// PKGPATH: gno.land/r/aib/bankertest
// SEND: 100ugnot,100uatone
package bankertest

import (
	"chain"
	"chain/banker"
	"chain/runtime"

	"gno.land/p/aib/limitedbanker"
	"gno.land/p/nt/testutils"
)

// multiple coins
func main() {
	var (
		banker = limitedbanker.New(
			banker.NewBanker(banker.BankerTypeRealmSend),
			chain.NewCoins(chain.NewCoin("uatone", 10), chain.NewCoin("ugnot", 10)),
		)
		from = runtime.CurrentRealm().Address()
		to   = testutils.TestAddress("test")
	)
	println("balance:", banker.GetCoins(from))
	println("spendable:", banker.SpendableCoins())

	banker.SendCoins(from, to, chain.NewCoins(chain.NewCoin("ugnot", 10)))

	println("balance:", banker.GetCoins(from))
	println("spendable:", banker.SpendableCoins())

	banker.SendCoins(from, to, chain.NewCoins(chain.NewCoin("uatone", 5)))

	println("balance:", banker.GetCoins(from))
	println("spendable:", banker.SpendableCoins())

	banker.SendCoins(from, to, chain.NewCoins(chain.NewCoin("uosmo", 5)))
}

// Output:
// balance: 100uatone,100ugnot
// spendable: 10uatone,10ugnot
// balance: 100uatone,90ugnot
// spendable: 10uatone,0ugnot
// balance: 95uatone,90ugnot
// spendable: 5uatone,0ugnot

// Error:
// spendLimit 5uatone,0ugnot exceeded for coins 5uosmo
