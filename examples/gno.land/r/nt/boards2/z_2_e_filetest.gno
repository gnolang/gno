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
	bid = boards2.CreateBoard("test123")
	tid = boards2.CreateThread(bid, "Foo", "bar")

	// Hide thread by flagging it so reply can't be submitted
	boards2.FlagThread(bid, tid, "reason")
}

func main() {
	boards2.CreateReply(bid, tid, 0, "Test reply")
}

// Error:
// thread with ID: 1 was hidden
