// PKGPATH: gno.land/r/test
package test

// A fully backed round trip at the int64 boundary: depositing MaxInt64 ugnot
// mints MaxInt64 wugnot, the realm's ugnot backing equals the supply at every
// step, and withdrawing all of it converts uint256 back to int64 exactly.

import (
	"chain"
	"chain/banker"
	"math"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gnoland/wugnot"
)

var (
	addr1 = testutils.TestAddress("test1")
	addrc = chain.PackageAddress("gno.land/r/gnoland/wugnot")
)

func main(cur realm) {
	testing.IssueCoins(addr1, chain.Coins{{"ugnot", math.MaxInt64}})
	report("start")

	coins := chain.Coins{{"ugnot", math.MaxInt64}}
	testing.SetOriginCaller(addr1)
	testing.SetOriginSend(coins)
	banker.NewBanker(banker.BankerTypeRealmSend, cur).SendCoins(addr1, addrc, coins)
	wugnot.Deposit(cross(cur))
	report("after deposit")

	wugnot.Withdraw(cross(cur), math.MaxInt64)
	report("after withdraw")
}

func report(step string) {
	rb := banker.NewReadonlyBanker()
	backing := rb.GetCoin(addrc, "ugnot")
	supply := wugnot.TotalSupply()
	println(step+":",
		"user ugnot", rb.GetCoin(addr1, "ugnot"),
		"user wugnot", wugnot.BalanceOf(addr1),
		"backing", backing,
		"supply", supply,
		"backed", uint256(backing) == supply)
}

// Output:
// start: user ugnot 9223372036854775807 user wugnot 0 backing 0 supply 0 backed true
// after deposit: user ugnot 0 user wugnot 9223372036854775807 backing 9223372036854775807 supply 9223372036854775807 backed true
// after withdraw: user ugnot 9223372036854775807 user wugnot 0 backing 0 supply 0 backed true
