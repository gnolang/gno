package main

import (
	"gno.land/p/demo/testutils"
	evaldao "gno.land/r/demo/evaluation_dao"
	"std"
)

func main() {
	// test1 -> test2
	test1 := testutils.TestAddress("test1")
	println("test1:", test1)
	test2 := testutils.TestAddress("test2")
	println("test2:", test2)
	test3 := testutils.TestAddress("test3")
	println("test3:", test3)
	println("============================")

	std.TestSetOrigCaller(test1)

	total := evaldao.TotalVotingPower()
	println("total voting power is :", total)
}

// Output:
