package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid boards2.BoardID
	tid boards2.PostID
	rid boards2.PostID
)

func init() {
	testing.SetOriginCaller(owner)
	bid = boards2.CreateBoard("test123")
	tid = boards2.CreateThread(bid, "foo", "bar")
	rid = boards2.CreateReply(bid, tid, 0, "reply")
}

func main() {
	// unfrozen thread - should work
	r2 := boards2.CreateReply(bid, tid, rid, "child 1")

	// frozen thread - can't reply
	boards2.FreezeReply(bid, tid, rid)
	boards2.CreateReply(bid, tid, r2, "child 2")
}

// Error:
// reply is frozen
