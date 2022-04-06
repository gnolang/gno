package main

// SEND: 100gnot

import (
	"gno.land/r/banktest"
	"std"
)

func main() {
	std.TestSetOrigPkgAddr(std.TestDerivePkgAddr("gno.land/r/banktest"))
	res := banktest.Deposit("gnot", 100)
	println(res)
	res = banktest.Render("")
	println(res)
}

// Output:
// returned!
// ## recent activity
//
//  * g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4 100gnot sent, 100gnot returned, at 1970-01-01 12:00am UTC
//
// ## total deposits
// 200gnot
