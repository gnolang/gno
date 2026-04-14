// PKGPATH: gno.land/r/test
package test

import (
	"chain"
	"chain/banker"
	"fmt"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gnoland/wugnot"
)

var (
	addr1 = testutils.TestAddress("test1")
	addrc = chain.PackageAddress("gno.land/r/gnoland/wugnot")
)

func main() {
	// issue ugnots
	testing.IssueCoins(addr1, chain.Coins{{"ugnot", 100000001}})
	printBalances()
	// println(wugnot.Render("queues"))
	// println("A -", wugnot.Render(""))

	// deposit of 123400ugnot from addr1
	// origin send must be simulated
	coins := chain.Coins{{"ugnot", 123_400}}
	testing.SetOriginCaller(addr1)
	testing.SetOriginSend(coins)
	banker.NewBanker(banker.BankerTypeRealmSend).SendCoins(addr1, addrc, coins)
	wugnot.Deposit(cross)
	printBalances()

	// withdraw of 4242ugnot to addr1
	wugnot.Withdraw(cross, 4242)
	printBalances()
}

func printBalances() {
	printSingleBalance := func(name string, addr address) {
		wugnotBal := wugnot.BalanceOf(addr)
		testing.SetOriginCaller(addr)
		robanker := banker.NewBanker(banker.BankerTypeReadonly)
		coins := robanker.GetCoins(addr).AmountOf("ugnot")
		fmt.Printf("| %-13s | addr=%s | wugnot=%-6d | ugnot=%-9d |\n",
			name, addr, wugnotBal, coins)
	}
	println("-----------")
	printSingleBalance("wugnot", addrc)
	printSingleBalance("addr1", addr1)
	println("-----------")
}

// Output:
// -----------
// | wugnot        | addr=g15vj5q08amlvyd0nx6zjgcvwq2d0gt9fcchrvum | wugnot=0      | ugnot=0         |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | wugnot=0      | ugnot=100000001 |
// -----------
// -----------
// | wugnot        | addr=g15vj5q08amlvyd0nx6zjgcvwq2d0gt9fcchrvum | wugnot=0      | ugnot=123400    |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | wugnot=123400 | ugnot=99876601  |
// -----------
// -----------
// | wugnot        | addr=g15vj5q08amlvyd0nx6zjgcvwq2d0gt9fcchrvum | wugnot=0      | ugnot=119158    |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | wugnot=119158 | ugnot=99880843  |
// -----------
