package main

import (
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	bid boards2.BoardID
	tid boards2.PostID
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)
	tid = boards2.CreateThread(cross, bid, "Foo", "bar")

	// Hide thread by flagging it so reply can't be submitted
	boards2.FlagThread(cross, bid, tid, "reason")
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.CreateReply(cross, bid, tid, 0, "Test reply")
}

// Error:
// thread is hidden
