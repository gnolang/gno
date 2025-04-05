package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	testing.SetOriginCaller(owner)
}

func main() {
	bid, found := boards2.GetBoardIDFromName("foobar")
	println(found)
	println(bid == 0)
}

// Output:
// false
// true
