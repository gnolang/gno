// PKGPATH: gno.land/r/demo/foo20_lp_test
package foo20_lp_test

import (
	"fmt"
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/foo20"
	"gno.land/r/demo/foo20_lp"
	"gno.land/r/demo/users"
)

var (
	addr1 = testutils.TestAddress("test1")
	addr2 = testutils.TestAddress("test2")
	addr3 = testutils.TestAddress("test3")
	addrc = std.DerivePkgAddr("gno.land/r/demo/foo20_lp")
	addrt = std.DerivePkgAddr("gno.land/r/demo/foo20_lp_test")
)

func main() {
	std.TestSetOrigPkgAddr(addrc)
	std.TestIssueCoins(addrc, std.Coins{{"ugnot", 100001337}}) // TODO: remove this

	// issue ugnots
	std.TestIssueCoins(addr1, std.Coins{{"ugnot", 100000001}})
	std.TestIssueCoins(addr2, std.Coins{{"ugnot", 100000002}})
	std.TestIssueCoins(addr3, std.Coins{{"ugnot", 100000003}})

	// issue foo20
	foo20Admin := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	std.TestSetOrigCaller(foo20Admin)
	foo20.Transfer(users.AddressOrName(addr1), 10001)
	foo20.Transfer(users.AddressOrName(addr2), 10002)
	foo20.Transfer(users.AddressOrName(addr3), 10003)

	// print initial state
	// printBalances()
	// println(foo20_lp.Render("queues"))
	// println("A -", foo20_lp.Render(""))

	// addr1 deposits ugnots
	std.TestSetOrigCaller(addr1)
	std.TestSetOrigSend(std.Coins{{"ugnot", 123_400}}, nil)
	foo20_lp.DepositUgnot()
	println("B -", foo20_lp.Render(""))

	// addr1 deposits ugnots again
	std.TestSetOrigCaller(addr1)
	std.TestSetOrigSend(std.Coins{{"ugnot", 567_800}}, nil)
	foo20_lp.DepositUgnot()
	println("C -", foo20_lp.Render(""))

	println(foo20_lp.Render("queues"))
	printBalances()
	return

	foo20_lp.DepositFoo20(10_001)
	foo20_lp.DepositFoo20(10_002)
	println("C -", foo20_lp.Render(""))
	foo20_lp.DepositFoo20(10_003)
	foo20_lp.DepositFoo20(10_004)
	foo20_lp.DepositFoo20(10_005)
	foo20_lp.DepositFoo20(10_006)
	println(foo20_lp.Render("queues"))
	println("D -", foo20_lp.Render(""))
	println("E -", foo20_lp.Render(""))
	foo20_lp.WithdrawFoo20(101)
	println("F -", foo20_lp.Render(""))
	foo20_lp.WithdrawFoo20(102)
	println("G -", foo20_lp.Render(""))
	foo20_lp.WithdrawUgnot(103)
	println("H -", foo20_lp.Render(""))
	println(foo20_lp.Render("queues"))
	println("--------------------------")
	printBalances()
}

func printBalances() {
	printSingleBalance := func(name string, addr std.Address) {
		foo20Bal := foo20.BalanceOf(users.AddressOrName(addr))
		std.TestSetOrigCaller(addr)
		abanker := std.GetBanker(std.BankerTypeOrigSend)
		acoins := abanker.GetCoins(addr).AmountOf("ugnot")
		bbanker := std.GetBanker(std.BankerTypeRealmIssue)
		bcoins := bbanker.GetCoins(addr).AmountOf("ugnot")
		cbanker := std.GetBanker(std.BankerTypeRealmSend)
		ccoins := cbanker.GetCoins(addr).AmountOf("ugnot")
		dbanker := std.GetBanker(std.BankerTypeReadonly)
		dcoins := dbanker.GetCoins(addr).AmountOf("ugnot")
		fmt.Printf("| %-13s | addr=%s | foo20=%-5d | ugnot=%-9d %-9d %-9d %-9d |\n",
			name, addr, foo20Bal, acoins, bcoins, ccoins, dcoins)
	}
	println("-----------")
	printSingleBalance("foo20_lp_test", addrt)
	printSingleBalance("foo20_lp", addrc)
	printSingleBalance("addr1", addr1)
	printSingleBalance("addr2", addr2)
	printSingleBalance("addr3", addr3)
	println("-----------")
}

// Output:
// @@@@@@@@@@@@@ g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 g12ddhtqgr5qsgy2keefnehrlwtwj09nner5jt3u 123400ugnot 123400
// not implemented: C
// B - LP is empty.
// @@@@@@@@@@@@@ g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 g12ddhtqgr5qsgy2keefnehrlwtwj09nner5jt3u 567800ugnot 567800
// not implemented: C
// C - LP is empty.
// ## Queues are empty
// -----------
// | foo20_lp_test | addr=g1erg29nwj2yu4ahl585369p2pxrfgdjjtha8zwm | foo20=0     | ugnot=200000000 200000000 200000000 200000000 |
// | foo20_lp      | addr=g12ddhtqgr5qsgy2keefnehrlwtwj09nner5jt3u | foo20=0     | ugnot=100001337 100001337 100001337 100001337 |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | foo20=10001 | ugnot=100000001 100000001 100000001 100000001 |
// | addr2         | addr=g1w3jhxapjta047h6lta047h6lta047h6laqcyu4 | foo20=10002 | ugnot=100000002 100000002 100000002 100000002 |
// | addr3         | addr=g1w3jhxapnta047h6lta047h6lta047h6lzfhfxt | foo20=10003 | ugnot=100000003 100000003 100000003 100000003 |
// -----------
