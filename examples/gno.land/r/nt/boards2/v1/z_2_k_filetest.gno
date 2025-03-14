package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner   = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	comment = "Second comment"
)

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
	tid = boards2.CreateThread(bid, "Foo", "bar")
	rid = boards2.CreateReply(bid, tid, 0, "Parent comment")

	// Flag parent post so it's hidden
	boards2.FlagReply(bid, tid, rid, "")
}

func main() {
	boards2.CreateReply(bid, tid, rid, "Sub comment")
}

// Error:
// replying to a hidden or frozen reply is not allowed
