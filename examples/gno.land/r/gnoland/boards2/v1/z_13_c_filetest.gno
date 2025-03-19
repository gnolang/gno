package main

import (
	"std"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var bid boards2.BoardID

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
}

func main() {
	boards2.FlagReply(bid, 404, 1, "")
}

// Error:
// thread does not exist with ID: 404
