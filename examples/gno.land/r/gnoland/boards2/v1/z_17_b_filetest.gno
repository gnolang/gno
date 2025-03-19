package main

import (
	"std"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var bid boards2.BoardID

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test123")
	boards2.FreezeBoard(bid)
}

func main() {
	boards2.FreezeBoard(bid)
}

// Error:
// board is frozen
