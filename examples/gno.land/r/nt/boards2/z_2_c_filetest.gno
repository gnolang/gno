package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	std.TestSetOrigCaller(owner)
	bid = boards2.CreateBoard("test123")
	tid = boards2.CreateThread(bid, "thread", "thread")
	rid = boards2.CreateReply(bid, tid, 0, "reply1")
	boards2.FlagReply(bid, tid, rid, "reason")
}

func main() {
	boards2.CreateReply(bid, tid, rid, "reply1.1")
}

// Error:
// reply with ID: 2 was hidden
