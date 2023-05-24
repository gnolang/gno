package main

import (
	"std"

	"gno.land/r/demo/banktest"
	"gno.land/p/demo/testutils"
)

func main() {
	banktestAddr := std.DerivePkgAddr("gno.land/r/banktest")

	// simulate a Deposit call.
	std.TestSetOrigPkgAddr(banktestAddr)
	testutils.TestOrigSend(banktestAddr, std.Coins{{"ugnot", 100000000}}, nil)
	res := banktest.Deposit("ugnot", 101000000)
	println(res)
}

// Error:
// cannot send "101000000ugnot", limit "100000000ugnot" exceeded with "" already spent
