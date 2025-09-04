package main

import (
	"chain"
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	bid boards2.BoardID
	tid boards2.PostID
	rid boards2.PostID
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)
	tid = boards2.CreateThread(cross, bid, "foo", "bar")
	rid = boards2.CreateReply(cross, bid, tid, 0, "reply")
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	// unfrozen thread - should work
	boards2.CreateReply(cross, bid, tid, rid, "child 1")

	// frozen thread - can't reply
	boards2.FreezeReply(cross, bid, tid, rid)
	boards2.CreateReply(cross, bid, tid, rid, "child 2")
}

// Error:
// replying to a hidden or frozen reply is not allowed
