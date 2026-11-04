// PKGPATH: gno.land/r/treasury/main

// A GRC20 treasury holds and pays amounts beyond the int64 range, and renders
// them in full.
package main

import (
	"chain"
	"chain/banker"
	"math"
	"testing"

	"gno.land/p/nt/grc20/v1"
	"gno.land/p/nt/treasury/v1"
)

func main(cur realm) {
	ownerAddr := chain.PackageAddress("gno.land/r/treasury/main")
	destAddr := chain.PackageAddress("gno.land/r/dest/main")

	token, ledger := grc20.NewToken("Big", "BIG", 18, 0, cur)
	if err := ledger.Mint(ownerAddr, math.MaxUint256); err != nil {
		panic(err)
	}
	lister := func() map[string]*grc20.Token { return map[string]*grc20.Token{"BIG": token} }

	gbanker, err := treasury.NewGRC20BankerWithOwner(ownerAddr, lister)
	if err != nil {
		panic(err)
	}
	cbanker, err := treasury.NewCoinsBankerWithOwner(ownerAddr, banker.NewBanker(banker.BankerTypeRealmSend, cur))
	if err != nil {
		panic(err)
	}
	testing.IssueCoins(ownerAddr, chain.NewCoins(chain.NewCoin("ugnot", math.MaxInt64)))

	t, err := treasury.New([]treasury.Banker{cbanker, gbanker}, "")
	if err != nil {
		panic(err)
	}

	// 2^200 is far beyond int64.
	payment := treasury.NewGRC20Payment("BIG", 1<<200, destAddr)
	println(payment.String())
	if err := t.Send(0, cur, payment); err != nil {
		panic(err)
	}
	println("dest:", token.BalanceOf(destAddr))
	println("owner:", token.BalanceOf(ownerAddr) == math.MaxUint256-1<<200)

	// More than the treasury holds is refused.
	err = t.Send(0, cur, treasury.NewGRC20Payment("BIG", math.MaxUint256, destAddr))
	println("overdraft:", err)

	// Coins balances are the int64 chain amounts, widened.
	coins, _ := t.Balances("Coins")
	println("coins:", coins[0].Amount == math.MaxInt64)

	println(t.Render("GRC20?history_size=0"))
}

// Output:
// 1606938044258990275541962092341162602522202993782792835301376BIG to g1fs92ep7z5vtrp26d767ag3zztsd6wqn9rew4ve
// dest: 1606938044258990275541962092341162602522202993782792835301376
// owner: true
// overdraft: failed to send payment: insufficient balance
// coins: true
// ## GRC20 Banker
// ### Address
// g1ynsdz5zaxhn9gnqtr6t40m5k4fueeutq7xy224
//
// ### Balances
// | Denom | Amount |
// | --- | --- |
// | BIG | 115792089237316193816632940749697632311307892324477961517254590225120294338559 |
