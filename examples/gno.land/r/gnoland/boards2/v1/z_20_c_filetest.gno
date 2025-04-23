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
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test123")
	tid = cross(boards2.CreateThread)(bid, "foo", "bar")
	rid = cross(boards2.CreateReply)(bid, tid, 0, "reply")

	testing.SetOriginCaller(owner)
}

func main() {
	// reply frozen
	println(boards2.IsReplyFrozen(bid, tid, rid))
	cross(boards2.FreezeReply)(bid, tid, rid)
	println(boards2.IsReplyFrozen(bid, tid, rid))
	cross(boards2.UnfreezeReply)(bid, tid, rid)
	println(boards2.IsReplyFrozen(bid, tid, rid))

	// thread frozen
	cross(boards2.FreezeThread)(bid, tid)
	println(boards2.IsReplyFrozen(bid, tid, rid))
	cross(boards2.UnfreezeThread)(bid, tid)
	println(boards2.IsReplyFrozen(bid, tid, rid))

	// board frozen
	cross(boards2.FreezeBoard)(bid)
	println(boards2.IsReplyFrozen(bid, tid, rid))
	cross(boards2.UnfreezeBoard)(bid)
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
