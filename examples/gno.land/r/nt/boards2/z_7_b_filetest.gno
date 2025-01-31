package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOrigCaller(owner)
}

func main() {
	bid, found := boards2.GetBoardIDFromName("foobar")
	println(found)
	println(bid == 0)
}

// Output:
// false
// true
