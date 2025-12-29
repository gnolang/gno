package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

var (
	bid      boards.ID
	rid, tid boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false, false)
	tid = boards2.CreateThread(cross, bid, "Foo", "bar")
	rid = boards2.CreateReply(cross, bid, tid, 0, "body")
	boards2.FlagReply(cross, bid, tid, rid, "")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.FlagReply(cross, bid, tid, rid, "")
}

// Error:
// flag count threshold exceeded: 1
