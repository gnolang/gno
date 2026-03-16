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

	rid = boards2.CreateReply(cross, bid, tid, 0, "Parent")
	boards2.CreateReply(cross, bid, tid, rid, "Child reply")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.DeleteReply(cross, bid, tid, rid)

	// Render content must contain the releted message instead of reply's body
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
// > ⚠ This comment has been deleted
// >
// > ↳ [Flag](/r/gnoland/boards2/v1:test-board/1/2/flag) • [Reply](/r/gnoland/boards2/v1:test-board/1/2/reply) [1] • [Edit](/r/gnoland/boards2/v1:test-board/1/2/edit) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteReply&boardID=1&replyID=2&threadID=1)
// >
// > >
// > > **[g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq)** `owner` · 2009-02-13 11:31pm UTC [#3](/r/gnoland/boards2/v1:test-board/1/3)
// > > Child reply
// > >
// > > ↳ [Flag](/r/gnoland/boards2/v1:test-board/1/3/flag) • [Reply](/r/gnoland/boards2/v1:test-board/1/3/reply) • [Edit](/r/gnoland/boards2/v1:test-board/1/3/edit) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteReply&boardID=1&replyID=3&threadID=1)
