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
)

func init() {
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
}

func main() {
	std.TestSetOrigCaller(addr1)

	printBalances()
	println("A -", foo20_lp.Render(""))
	println(foo20_lp.Render("queues"))
	foo20_lp.DepositFoo20(10_000)
	println("B -", foo20_lp.Render(""))
	foo20_lp.DepositFoo20(10_000)
	println("C -", foo20_lp.Render(""))
	foo20_lp.DepositFoo20(10_000)
	foo20_lp.DepositFoo20(10_000)
	foo20_lp.DepositFoo20(10_000)
	foo20_lp.DepositFoo20(10_000)
	println(foo20_lp.Render("queues"))
	println("D -", foo20_lp.Render(""))
	std.TestSetOrigSend(std.Coins{{"ugnot", 123_400}}, nil)
	foo20_lp.DepositUgnot()
	println("E -", foo20_lp.Render(""))
	foo20_lp.WithdrawFoo20(100)
	println("F -", foo20_lp.Render(""))
	foo20_lp.WithdrawFoo20(100)
	println("G -", foo20_lp.Render(""))
	foo20_lp.WithdrawUgnot(100)
	println("H -", foo20_lp.Render(""))
	println(foo20_lp.Render("queues"))
	println("--------------------------")
	printBalances()
}

func printBalances() {
	printSingleBalance := func(name string, addr std.Address) {
		foo20Bal := foo20.BalanceOf(users.AddressOrName(addr))
		ugnotBal := 0
		fmt.Printf("%s=%s foo20=%d ugnot=%d\n", name, addr, foo20Bal, ugnotBal)
		{
			banker := std.GetBanker(std.BankerTypeOrigSend)
			coins := banker.GetCoins(addr)
			println(coins)
		}
		{
			banker := std.GetBanker(std.BankerTypeRealmIssue)
			coins := banker.GetCoins(addr)
			println(coins)
		}
		{
			banker := std.GetBanker(std.BankerTypeRealmSend)
			coins := banker.GetCoins(addr)
			println(coins)
		}
		{
			banker := std.GetBanker(std.BankerTypeReadonly)
			coins := banker.GetCoins(addr)
			println(coins)
		}
	}
	printSingleBalance("contract", addrc)
	printSingleBalance("   addr1", addr1)
	printSingleBalance("   addr2", addr2)
	printSingleBalance("   addr3", addr3)
}

// Output:
// contract=g12ddhtqgr5qsgy2keefnehrlwtwj09nner5jt3u foo20=0 ugnot=0
//
//
//
//
//    addr1=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 foo20=10001 ugnot=0
// 100000001ugnot
// 100000001ugnot
// 100000001ugnot
// 100000001ugnot
//    addr2=g1w3jhxapjta047h6lta047h6lta047h6laqcyu4 foo20=10002 ugnot=0
// 100000002ugnot
// 100000002ugnot
// 100000002ugnot
// 100000002ugnot
//    addr3=g1w3jhxapnta047h6lta047h6lta047h6lzfhfxt foo20=10003 ugnot=0
// 100000003ugnot
// 100000003ugnot
// 100000003ugnot
// 100000003ugnot
// A - LP is empty.
// ## Queues are empty
// B - LP has a total of 10000 foo20 in 1 orders for a capacity of 1000000 ugnot.
// C - LP has a total of 20000 foo20 in 2 orders for a capacity of 2000000 ugnot.
// ## foo20Queue (total=60000, orders=6)
// - addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 amount=10000 (cum=10000)
// - addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 amount=10000 (cum=20000)
// - addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 amount=10000 (cum=30000)
// - addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 amount=10000 (cum=40000)
// - addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 amount=10000 (cum=50000)
// - addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 amount=10000 (cum=60000)
//
// D - LP has a total of 60000 foo20 in 6 orders for a capacity of 6000000 ugnot.
// not implemented: C
// E - LP has a total of 60000 foo20 in 6 orders for a capacity of 6000000 ugnot.
// not implemented: D
// F - LP has a total of 60000 foo20 in 6 orders for a capacity of 6000000 ugnot.
// not implemented: D
// G - LP has a total of 60000 foo20 in 6 orders for a capacity of 6000000 ugnot.
// not implemented: E
// H - LP has a total of 60000 foo20 in 6 orders for a capacity of 6000000 ugnot.
// ## foo20Queue (total=60000, orders=6)
// - addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 amount=10000 (cum=10000)
// - addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 amount=10000 (cum=20000)
// - addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 amount=10000 (cum=30000)
// - addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 amount=10000 (cum=40000)
// - addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 amount=10000 (cum=50000)
// - addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 amount=10000 (cum=60000)
//
// --------------------------
// contract=g12ddhtqgr5qsgy2keefnehrlwtwj09nner5jt3u foo20=0 ugnot=0
//
//
//
//
//    addr1=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 foo20=10001 ugnot=0
// 100000001ugnot
// 100000001ugnot
// 100000001ugnot
// 100000001ugnot
//    addr2=g1w3jhxapjta047h6lta047h6lta047h6laqcyu4 foo20=10002 ugnot=0
// 100000002ugnot
// 100000002ugnot
// 100000002ugnot
// 100000002ugnot
//    addr3=g1w3jhxapnta047h6lta047h6lta047h6lzfhfxt foo20=10003 ugnot=0
// 100000003ugnot
// 100000003ugnot
// 100000003ugnot
// 100000003ugnot
