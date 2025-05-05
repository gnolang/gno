// PKGPATH: gno.land/r/demo/wugnot_test
package wugnot_test

import (
	"fmt"
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/wugnot"
)

var (
	addr1 = testutils.TestAddress("test1")
	addrc = std.DerivePkgAddr("gno.land/r/demo/wugnot")
)

func main() {
	// issue ugnots
	testing.IssueCoins(addr1, std.Coins{{"ugnot", 100000001}})
	printBalances()
	// println(wugnot.Render("queues"))
	// println("A -", wugnot.Render(""))

	// deposit of 123400ugnot from addr1
	// origin send must be simulated
	coins := std.Coins{{"ugnot", 123_400}}
	testing.SetOriginCaller(addr1)
	testing.SetOriginSend(coins)
	std.NewBanker(std.BankerTypeRealmSend).SendCoins(addr1, addrc, coins)
	cross(wugnot.Deposit)()
	printBalances()

	// withdraw of 4242ugnot to addr1
	cross(wugnot.Withdraw)(4242)
	printBalances()
}

func printBalances() {
	printSingleBalance := func(name string, addr std.Address) {
		wugnotBal := wugnot.BalanceOf(addr)
		testing.SetOriginCaller(addr)
		robanker := std.NewBanker(std.BankerTypeReadonly)
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
// | wugnot        | addr=g1pf6dv9fjk3rn0m4jjcne306ga4he3mzmupfjl6 | wugnot=0      | ugnot=0         |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | wugnot=0      | ugnot=100000001 |
// -----------
// -----------
// | wugnot        | addr=g1pf6dv9fjk3rn0m4jjcne306ga4he3mzmupfjl6 | wugnot=0      | ugnot=123400    |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | wugnot=123400 | ugnot=99876601  |
// -----------
// -----------
// | wugnot        | addr=g1pf6dv9fjk3rn0m4jjcne306ga4he3mzmupfjl6 | wugnot=0      | ugnot=119158    |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | wugnot=119158 | ugnot=99880843  |
// -----------
