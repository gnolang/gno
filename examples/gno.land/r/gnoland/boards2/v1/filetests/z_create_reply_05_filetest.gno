package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

var (
	bid      boards.ID
	tid, rid boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false, false)
	tid = boards2.CreateThread(cross, bid, "Foo", "bar")
	rid = boards2.CreateReply(cross, bid, tid, 0, "reply1")

	// Hide thread by flagging it so reply of a reply can't be submitted
	boards2.FlagThread(cross, bid, tid, "reason")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateReply(cross, bid, tid, rid, "reply1.1")
}

// Error:
// thread is hidden
