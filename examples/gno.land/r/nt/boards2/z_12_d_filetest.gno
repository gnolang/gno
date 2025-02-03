package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var bid boards2.BoardID

func init() {
	std.TestSetOrigCaller(owner)
	bid = boards2.CreateBoard("test-board")
}

func main() {
	boards2.EditReply(bid, 404, 0, "body")
}

// Error:
// thread does not exist with ID: 404
