// PKGPATH: gno.land/r/test
package test

// Withdraw takes a uint256, but every wugnot comes from deposited ugnot, so an
// amount beyond the int64 range can never be backed and is refused.

import (
	"chain"
	"chain/banker"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gnoland/wugnot"
)

var (
	addr1 = testutils.TestAddress("test1")
	addrc = chain.PackageAddress("gno.land/r/gnoland/wugnot")
)

func main(cur realm) {
	testing.IssueCoins(addr1, chain.Coins{{"ugnot", 10_000}})
	coins := chain.Coins{{"ugnot", 5_000}}
	testing.SetOriginCaller(addr1)
	testing.SetOriginSend(coins)
	banker.NewBanker(banker.BankerTypeRealmSend, cur).SendCoins(addr1, addrc, coins)
	wugnot.Deposit(cross(cur))
	println("balance:", wugnot.BalanceOf(addr1))
	println("supply:", wugnot.TotalSupply())

	wugnot.Withdraw(cross(cur), 1<<64)
}

// Output:
// balance: 5000
// supply: 5000

// Error:
// Insufficient balance: 5000 available, 18446744073709551616 needed.
