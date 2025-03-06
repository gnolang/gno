package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid boards2.BoardID
	tid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test123")
	tid = boards2.CreateThread(bid, "foo", "bar")
	boards2.FreezeThread(bid, tid)
}

func main() {
	// cannot reply to a frozen thread
	boards2.CreateReply(bid, tid, 0, "foobar")
}

// Error:
// thread is frozen
