package main

import (
	"std"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name  = "test123"
)

var bid boards2.BoardID

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard(name)
}

func main() {
	bid2, found := boards2.GetBoardIDFromName(name)
	println(found)
	println(bid2 == bid)
}

// Output:
// true
// true
