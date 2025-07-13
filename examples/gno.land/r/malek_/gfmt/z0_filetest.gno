// PKGPATH: gno.land/r/malek_/gfmt_test
package gfmt_test

import (
	"fmt"
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/malek_/gfmt"

	pusers "gno.land/p/demo/users"
)

var (
	addr1 = testutils.TestAddress("test1")
	addrc = std.DerivePkgAddr("gno.land/r/malek_/gfmt")
	addrt = std.DerivePkgAddr("gno.land/r/malek_/gfmt_test")
)

func main() {
	std.TestSetOrigPkgAddr(addrc)
	std.TestIssueCoins(addrc, std.Coins{{"ugnot", 100000001}}) // TODO: remove this

	// issue ugnots
	std.TestIssueCoins(addr1, std.Coins{{"ugnot", 100000001}})

	// print initial state
	printBalances()
	// println(gfmt.Render("queues"))
	// println("A -", gfmt.Render(""))

	std.TestSetOrigCaller(addr1)
	std.TestSetOrigSend(std.Coins{{"ugnot", 1_000_000}}, nil)
	gfmt.Deposit()
	printBalances()
	gfmt.Withdraw(4242)
	printBalances()
}

func printBalances() {
	printSingleBalance := func(name string, addr std.Address) {
		gfmtBal := gfmt.BalanceOf(pusers.AddressOrName(addr))
		std.TestSetOrigCaller(addr)
		robanker := std.GetBanker(std.BankerTypeReadonly)
		coins := robanker.GetCoins(addr).AmountOf("ugnot")
		fmt.Printf("| %-13s | addr=%s | gfmt=%-7d | ugnot=%-9d |\n",
			name, addr, gfmtBal, coins)
	}
	println("-----------")
	printSingleBalance("gfmt_test", addrt)
	printSingleBalance("gfmt", addrc)
	printSingleBalance("addr1", addr1)
	println("-----------")
}

// Output:
// -----------
// | gfmt_test     | addr=g15pyg9fhp8nlt8vhvh4pwhvamdxzuk052mlj5s7 | gfmt=0       | ugnot=200000000 |
// | gfmt          | addr=g1hralrm7msf9uc3nzjmpvmsxapfdxx2j5k4yhjm | gfmt=0       | ugnot=100000001 |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | gfmt=0       | ugnot=100000001 |
// -----------
// -----------
// | gfmt_test     | addr=g15pyg9fhp8nlt8vhvh4pwhvamdxzuk052mlj5s7 | gfmt=1000000 | ugnot=200000000 |
// | gfmt          | addr=g1hralrm7msf9uc3nzjmpvmsxapfdxx2j5k4yhjm | gfmt=0       | ugnot=100000001 |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | gfmt=0       | ugnot=100000001 |
// -----------
// -----------
// | gfmt_test     | addr=g15pyg9fhp8nlt8vhvh4pwhvamdxzuk052mlj5s7 | gfmt=995758  | ugnot=200004242 |
// | gfmt          | addr=g1hralrm7msf9uc3nzjmpvmsxapfdxx2j5k4yhjm | gfmt=0       | ugnot=99995759  |
// | addr1         | addr=g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 | gfmt=0       | ugnot=100000001 |
// -----------
