package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	moderator address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var (
	bid      boards.ID
	rid, tid boards.ID
)

func init() {
	// Created a board with flagging threshold greater than 1
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false, false)
	boards2.SetFlaggingThreshold(cross, bid, 2)

	// Create a reply so the realm owner can flag and hide it with a single flag
	tid = boards2.CreateThread(cross, bid, "Foo", "bar")
	rid = boards2.CreateReply(cross, bid, tid, 0, "body")
	// Also freeze board to make sure that realm owner can still flag the reply
	boards2.FreezeBoard(cross, bid)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.FlagReply(cross, bid, tid, rid, "")

	// Render content must contain a message about the hidden reply
	println(boards2.Render("test-board/1/2"))
}

// Output:
// ## Foo
//
// **[g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq)** `owner` · 2009-02-13 11:31pm UTC
// bar
//
// ↳ [Flag](/r/gnoland/boards2/v1:test-board/1/flag) • [Repost](/r/gnoland/boards2/v1:test-board/1/repost) • [Show all Replies](/r/gnoland/boards2/v1:test-board/1)
//
//
// >
// > **[g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq)** `owner` · 2009-02-13 11:31pm UTC [#2](/r/gnoland/boards2/v1:test-board/1/2)
// > ⚠ Reply is hidden as it has been flagged as [inappropriate](/r/gnoland/boards2/v1:test-board/1/2/flagging-reasons)
