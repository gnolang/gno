// PKGPATH: gno.land/r/treasury/main

package main

import (
	"chain"
	"chain/banker"
	"testing"

	"gno.land/p/nt/treasury/v0"
)

// coins is this realm's own slice. It is handed to NewCoinsPayment and then
// rewritten, which is what a realm supplying the coins for a governance
// payment proposal can do between the vote and the execution.
var coins = chain.Coins{chain.Coin{Denom: "ugnot", Amount: 10}}

func main(cur realm) {
	ownerAddr := chain.PackageAddress("gno.land/r/treasury/main")
	destAddr := chain.PackageAddress("gno.land/r/dest/main")

	banker_ := banker.NewBanker(banker.BankerTypeRealmSend, cur)
	cbanker, err := treasury.NewCoinsBankerWithOwner(ownerAddr, banker_)
	if err != nil {
		panic("failed to create CoinsBanker: " + err.Error())
	}
	testing.IssueCoins(ownerAddr, chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	payment := treasury.NewCoinsPayment(coins, destAddr)
	println("payment as created:", payment.String())

	// The supplying realm rewrites its own array. The payment must not follow.
	coins[0].Amount = 900
	println("payment after the rewrite:", payment.String())

	if err := cbanker.Send(0, cur, payment); err != nil {
		panic("failed to send payment: " + err.Error())
	}
	println("owner balance after send:", cbanker.Balances()[0].Amount)
}

// Output:
// payment as created: 10ugnot to g1fs92ep7z5vtrp26d767ag3zztsd6wqn9rew4ve
// payment after the rewrite: 10ugnot to g1fs92ep7z5vtrp26d767ag3zztsd6wqn9rew4ve
// owner balance after send: 990
