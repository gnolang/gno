package main

// SEND: 100gnot

import (
	"gno.land/r/banktest"
	"std"
)

func main() {
	std.TestSetOrigPkgAddr(std.TestDerivePkgAddr("gno.land/r/banktest"))
	res := banktest.Deposit("gnot", 101)
	println(res)
}

// Error:
// cannot send 101gnot
