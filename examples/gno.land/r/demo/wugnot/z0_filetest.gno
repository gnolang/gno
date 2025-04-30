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
	addrt = std.DerivePkgAddr("gno.land/r/demo/wugnot_test")
)

func main() {
	testing.IssueCoins(addrc, std.Coins{{"ugnot", 100000001}}) // TODO: remove this

	// issue ugnots
	testing.IssueCoins(addr1, std.Coins{{"ugnot", 100000001}})

	// print initial state
	printBalances()
	// println(wugnot.Render("queues"))
	// println("A -", wugnot.Render(""))

	testing.SetOriginCaller(addr1)
	testing.SetOriginSend(std.Coins{{"ugnot", 123_400}})
	cross(wugnot.Deposit)()
	printBalances()
	cross(wugnot.Withdraw)(4242)
	printBalances()
}

func printBalances() {
	printSingleBalance := func(name string, addr std.Address) {
		wugnotBal := wugnot.BalanceOf(addr)
		testing.SetOriginCaller(addr)
		robanker := std.NewBanker(std.BankerTypeReadonly)
		coins := robanker.GetCoins(addr).AmountOf("ugnot")
		fmt.Printf("| %-13s | addr=%s | wugnot=%-5d | ugnot=%-9d |\n",
			name, addr, wugnotBal, coins)
	}
	println("-----------")
	printSingleBalance("wugnot_test", addrt)
	printSingleBalance("wugnot", addrc)
	printSingleBalance("addr1", addr1)
	println("-----------")
}

// Output:
// -----------
// | wugnot_test   | addr=g19rmydykafrqyyegc8uuaxxpzqwzcnxraj2dev9 | wugnot=0     | ugnot=0         |
// | wugnot        | addr=g1pf6dv9fjk3rn0m4jjcne306ga4he3mzmupfjl6 | wugnot=0     | ugnot=100000001 |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | wugnot=0     | ugnot=100000001 |
// -----------
// -----------
// | wugnot_test   | addr=g19rmydykafrqyyegc8uuaxxpzqwzcnxraj2dev9 | wugnot=123400 | ugnot=0         |
// | wugnot        | addr=g1pf6dv9fjk3rn0m4jjcne306ga4he3mzmupfjl6 | wugnot=0     | ugnot=100000001 |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | wugnot=0     | ugnot=100000001 |
// -----------
// -----------
// | wugnot_test   | addr=g19rmydykafrqyyegc8uuaxxpzqwzcnxraj2dev9 | wugnot=119158 | ugnot=4242      |
// | wugnot        | addr=g1pf6dv9fjk3rn0m4jjcne306ga4he3mzmupfjl6 | wugnot=0     | ugnot=99995759  |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | wugnot=0     | ugnot=100000001 |
// -----------
