package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test123")
	tid = boards2.CreateThread(bid, "Foo", "bar")
	rid = boards2.CreateReply(bid, tid, 0, "reply1")

	// Hide thread by flagging it so reply of a reply can't be submitted
	boards2.FlagThread(bid, tid, "reason")
}

func main() {
	boards2.CreateReply(bid, tid, rid, "reply1.1")
}

// Error:
// thread with ID: 1 was hidden
