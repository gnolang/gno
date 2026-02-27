// Render comment/reply view of a flagged comment.
package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	boardName         = "test-board"
)

var (
	threadID boards.ID
	replyID  boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a board and a thread
	boardID := boards2.CreateBoard(cross, boardName, false, false)
	threadID = boards2.CreateThread(cross, boardID, "Foo", "Body")

	// Create a comments with a reply
	replyID = boards2.CreateReply(cross, boardID, threadID, 0, "Second comment")
	boards2.CreateReply(cross, boardID, threadID, replyID, "Third comment")

	// Flag the comment
	boards2.FlagReply(cross, boardID, threadID, replyID, "Reason")
}

func main() {
	path := boardName + "/" + threadID.String() + "/" + replyID.String()
	println(boards2.Render(path))
}

// Output:
// ## Foo
//
// **[g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq)** `owner` · 2009-02-13 11:31pm UTC
// Body
//
// ↳ [Flag](/r/gnoland/boards2/v1:test-board/1/flag) • [Repost](/r/gnoland/boards2/v1:test-board/1/repost) • [Comment](/r/gnoland/boards2/v1:test-board/1/reply) [1] • [Edit](/r/gnoland/boards2/v1:test-board/1/edit) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteThread&boardID=1&threadID=1) • [Show all Replies](/r/gnoland/boards2/v1:test-board/1)
//
//
// >
// > **[g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq)** `owner` · 2009-02-13 11:31pm UTC [#2](/r/gnoland/boards2/v1:test-board/1/2)
// > ⚠ Reply is hidden as it has been flagged as [inappropriate](/r/gnoland/boards2/v1:test-board/1/2/flagging-reasons)
// >
// > >
// > > **[g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq)** `owner` · 2009-02-13 11:31pm UTC [#3](/r/gnoland/boards2/v1:test-board/1/3)
// > > Third comment
// > >
// > > ↳ [Flag](/r/gnoland/boards2/v1:test-board/1/3/flag) • [Reply](/r/gnoland/boards2/v1:test-board/1/3/reply) • [Edit](/r/gnoland/boards2/v1:test-board/1/3/edit) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteReply&boardID=1&replyID=3&threadID=1)
