package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	bid boards2.BoardID
	tid boards2.PostID
	rid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)
	tid = boards2.CreateThread(cross, bid, "foo", "bar")
	rid = boards2.CreateReply(cross, bid, tid, 0, "reply")
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	// reply frozen
	println(boards2.IsReplyFrozen(bid, tid, rid))
	boards2.FreezeReply(cross, bid, tid, rid)
	println(boards2.IsReplyFrozen(bid, tid, rid))
	boards2.UnfreezeReply(cross, bid, tid, rid)
	println(boards2.IsReplyFrozen(bid, tid, rid))

	// thread frozen
	boards2.FreezeThread(cross, bid, tid)
	println(boards2.IsReplyFrozen(bid, tid, rid))
	boards2.UnfreezeThread(cross, bid, tid)
	println(boards2.IsReplyFrozen(bid, tid, rid))

	// board frozen
	boards2.FreezeBoard(cross, bid)
	println(boards2.IsReplyFrozen(bid, tid, rid))
	boards2.UnfreezeBoard(cross, bid)
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
