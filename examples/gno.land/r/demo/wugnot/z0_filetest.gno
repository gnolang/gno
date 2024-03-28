// PKGPATH: gno.land/r/demo/wugnot_test
package wugnot_test

import (
	"fmt"
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/wugnot"

	pusers "gno.land/p/demo/users"
)

var (
	addr1 = testutils.TestAddress("test1")
	addrc = std.DerivePkgAddr("gno.land/r/demo/wugnot")
	addrt = std.DerivePkgAddr("gno.land/r/demo/wugnot_test")
)

func main() {
	std.TestSetOrigPkgAddr(addrc)
	std.TestIssueCoins(addrc, std.Coins{{"ugnot", 100000001}}) // TODO: remove this

	// issue ugnots
	std.TestIssueCoins(addr1, std.Coins{{"ugnot", 100000001}})

	// print initial state
	printBalances()
	// println(wugnot.Render("queues"))
	// println("A -", wugnot.Render(""))

	std.TestSetOrigCaller(addr1)
	std.TestSetOrigSend(std.Coins{{"ugnot", 123_400}}, nil)
	wugnot.Deposit()
	printBalances()
	wugnot.Withdraw(4242)
	printBalances()
}

func printBalances() {
	printSingleBalance := func(name string, addr std.Address) {
		wugnotBal := wugnot.BalanceOf(pusers.AddressOrName(addr))
		std.TestSetOrigCaller(addr)
		abanker := std.GetBanker(std.BankerTypeOrigSend)
		acoins := abanker.GetCoins(addr).AmountOf("ugnot")
		bbanker := std.GetBanker(std.BankerTypeRealmIssue)
		bcoins := bbanker.GetCoins(addr).AmountOf("ugnot")
		cbanker := std.GetBanker(std.BankerTypeRealmSend)
		ccoins := cbanker.GetCoins(addr).AmountOf("ugnot")
		dbanker := std.GetBanker(std.BankerTypeReadonly)
		dcoins := dbanker.GetCoins(addr).AmountOf("ugnot")
		fmt.Printf("| %-13s | addr=%s | wugnot=%-5d | ugnot=%-9d %-9d %-9d %-9d |\n",
			name, addr, wugnotBal, acoins, bcoins, ccoins, dcoins)
	}
	println("-----------")
	printSingleBalance("wugnot_test", addrt)
	printSingleBalance("wugnot", addrc)
	printSingleBalance("addr1", addr1)
	println("-----------")
}

// Output:
// -----------
// | wugnot_test   | addr=g19rmydykafrqyyegc8uuaxxpzqwzcnxraj2dev9 | wugnot=0     | ugnot=200000000 200000000 200000000 200000000 |
// | wugnot        | addr=g1pf6dv9fjk3rn0m4jjcne306ga4he3mzmupfjl6 | wugnot=0     | ugnot=100000001 100000001 100000001 100000001 |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | wugnot=0     | ugnot=100000001 100000001 100000001 100000001 |
// -----------
// -----------
// | wugnot_test   | addr=g19rmydykafrqyyegc8uuaxxpzqwzcnxraj2dev9 | wugnot=123400 | ugnot=200000000 200000000 200000000 200000000 |
// | wugnot        | addr=g1pf6dv9fjk3rn0m4jjcne306ga4he3mzmupfjl6 | wugnot=0     | ugnot=100000001 100000001 100000001 100000001 |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | wugnot=0     | ugnot=100000001 100000001 100000001 100000001 |
// -----------
// -----------
// | wugnot_test   | addr=g19rmydykafrqyyegc8uuaxxpzqwzcnxraj2dev9 | wugnot=119158 | ugnot=200004242 200004242 200004242 200004242 |
// | wugnot        | addr=g1pf6dv9fjk3rn0m4jjcne306ga4he3mzmupfjl6 | wugnot=0     | ugnot=99995759  99995759  99995759  99995759  |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | wugnot=0     | ugnot=100000001 100000001 100000001 100000001 |
// -----------
