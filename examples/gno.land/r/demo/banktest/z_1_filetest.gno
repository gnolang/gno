package main

import (
	"std"

	"gno.land/r/demo/banktest"
)

func main() {
	banktestAddr := std.DerivePkgAddr("gno.land/r/demo/banktest")

	// simulate a Deposit call.
	std.TestIssueCoins(banktestAddr, std.Coins{{"ugnot", 100000000}})
	std.TestSetOrigSend(std.Coins{{"ugnot", 100000000}}, nil)
	res := banktest.Deposit("ugnot", 101000000)
	println(res)
}

// Error:
// cannot send "101000000ugnot", limit "100000000ugnot" exceeded with "" already spent
