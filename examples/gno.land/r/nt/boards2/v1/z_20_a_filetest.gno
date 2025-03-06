package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid boards2.BoardID
	tid boards2.PostID
	rid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test123")
	tid = boards2.CreateThread(bid, "foo", "bar", false)
	rid = boards2.CreateReply(bid, tid, 0, "reply")

	boards2.FreezeThread(bid, tid)
}

func main() {
	// Attempt to freeze a reply on frozen thread
	boards2.FreezeReply(bid, tid, rid)
}

// Error:
// thread is frozen
