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
	bid = boards2.CreateBoard(cross, "test-board", false, false)
	tid = boards2.CreateThread(cross, bid, "Foo", "bar")
	rid = boards2.CreateReply(cross, bid, tid, 0, "body")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.DeleteReply(cross, bid, tid, rid)

	// Ensure reply doesn't exist
	println(boards2.Render("test-board/1/2"))
}

// Output:
// Reply not found
