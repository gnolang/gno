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
	tid = boards2.CreateThread(bid, "thread", "thread")
	rid = boards2.CreateReply(bid, tid, 0, "reply1")

	// Hide reply by flagging it so sub reply can't be submitted
	boards2.FlagReply(bid, tid, rid, "reason")
}

func main() {
	boards2.CreateReply(bid, tid, rid, "reply1.1")
}

// Error:
// replying to a hidden reply is not allowed
