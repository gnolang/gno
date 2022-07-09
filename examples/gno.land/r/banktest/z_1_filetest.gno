package main

import (
	"std"

	"gno.land/r/banktest"
)

func main() {
	banktestAddr := std.TestDerivePkgAddr("gno.land/r/banktest")

	// simulate a Deposit call.
	std.TestSetOrigPkgAddr(banktestAddr)
	std.TestIssueCoins(banktestAddr, std.Coins{{"ugnot", 100}})
	std.TestSetOrigSend(std.Coins{{"ugnot", 100}}, nil)
	res := banktest.Deposit("ugnot", 101)
	println(res)
}

// Error:
// cannot send "101ugnot"
