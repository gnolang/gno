package main

import (
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	comment = "Second comment"
)

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false)
	tid = boards2.CreateThread(cross, bid, "Foo", "bar")
	rid = boards2.CreateReply(cross, bid, tid, 0, "Parent comment")

	// Flag parent post so it's hidden
	boards2.FlagReply(cross, bid, tid, rid, "")
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.CreateReply(cross, bid, tid, rid, "Sub comment")
}

// Error:
// replying to a hidden or frozen reply is not allowed
