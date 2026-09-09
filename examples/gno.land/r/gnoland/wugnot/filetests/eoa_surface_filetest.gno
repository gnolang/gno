// PKGPATH: gno.land/r/test/wugnotsurface

package wugnotsurface

import (
	"chain"
	"chain/banker"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gnoland/wugnot"
)

func main(cur realm) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carol := testutils.TestAddress("carol")

	testing.IssueCoins(alice, chain.Coins{{"ugnot", 1_000}})
	testing.SetOriginCaller(alice)
	testing.SetOriginSend(chain.Coins{{"ugnot", 1_000}})
	wugnotAddr := chain.PackageAddress("gno.land/r/gnoland/wugnot")
	banker.NewBanker(banker.BankerTypeRealmSend, cur).SendCoins(alice, wugnotAddr, chain.Coins{{"ugnot", 1_000}})
	wugnot.Deposit(cross(cur))

	testing.SetRealm(testing.NewUserRealm(alice))
	wugnot.Approve(cross(cur), bob, 400)

	testing.SetRealm(testing.NewUserRealm(bob))
	wugnot.TransferFrom(cross(cur), alice, carol, 250)

	testing.SetRealm(testing.NewUserRealm(alice))
	wugnot.Transfer(cross(cur), carol, 100)

	// An intermediate realm is the immediate spender. It cannot consume the
	// allowance granted to bob, even though alice signed the transaction.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/wugnotsurface"))
	err := revive(func() {
		wugnot.TransferFrom(cross(cur), alice, cur.Address(), 1)
	})
	println(err)
	println(wugnot.BalanceOf(alice), wugnot.BalanceOf(carol), wugnot.Allowance(alice, bob))
}

// Output:
// insufficient allowance
// 650 350 150
