package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOriginCaller(owner)
}

func main() {
	bid, found := boards2.GetBoardIDFromName("foobar")
	println(found)
	println(bid == 0)
}

// Output:
// false
// true
