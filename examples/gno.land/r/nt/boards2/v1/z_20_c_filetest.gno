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
	tid = boards2.CreateThread(bid, "foo", "bar")
	rid = boards2.CreateReply(bid, tid, 0, "reply")
}

func main() {
	// reply frozen
	println(boards2.IsReplyFrozen(bid, tid, rid))
	boards2.FreezeReply(bid, tid, rid)
	println(boards2.IsReplyFrozen(bid, tid, rid))
	boards2.UnfreezeReply(bid, tid, rid)
	println(boards2.IsReplyFrozen(bid, tid, rid))

	// thread frozen
	boards2.FreezeThread(bid, tid)
	println(boards2.IsReplyFrozen(bid, tid, rid))
	boards2.UnfreezeThread(bid, tid)
	println(boards2.IsReplyFrozen(bid, tid, rid))

	// board frozen
	boards2.FreezeBoard(bid)
	println(boards2.IsReplyFrozen(bid, tid, rid))
	boards2.UnfreezeBoard(bid)
	println(boards2.IsReplyFrozen(bid, tid, rid))
}

// Output:
// false
// true
// false
// true
// false
// true
// false
