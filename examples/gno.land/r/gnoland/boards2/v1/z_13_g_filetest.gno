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
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false, false)
	tid = boards2.CreateThread(cross, bid, "Foo", "bar")
	rid = boards2.CreateReply(cross, bid, tid, 0, "body")

	// Invite a member using a role with permission to flag replies
	boards2.InviteMember(cross, bid, moderator, boards2.RoleModerator)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(moderator))

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
// ↳ [Flag](/r/gnoland/boards2/v1:test-board/1/flag) • [Repost](/r/gnoland/boards2/v1:test-board/1/repost) • [Comment](/r/gnoland/boards2/v1:test-board/1/reply) [1] • [Edit](/r/gnoland/boards2/v1:test-board/1/edit) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteThread&boardID=1&threadID=1) • [Show all Replies](/r/gnoland/boards2/v1:test-board/1)
//
//
// >
// > **[g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq)** `owner` · 2009-02-13 11:31pm UTC [#2](/r/gnoland/boards2/v1:test-board/1/2)
// > ⚠ Reply is hidden as it has been flagged as [inappropriate](/r/gnoland/boards2/v1:test-board/1/2/flagging-reasons)
