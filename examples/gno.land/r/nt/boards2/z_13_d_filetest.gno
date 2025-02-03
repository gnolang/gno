package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid boards2.BoardID
	tid boards2.PostID
)

func init() {
	std.TestSetOrigCaller(owner)
	bid = boards2.CreateBoard("test-board")
	tid = boards2.CreateThread(bid, "Foo", "bar")
}

func main() {
	boards2.FlagReply(bid, tid, 404, "")
}

// Error:
// reply does not exist with ID: 404
