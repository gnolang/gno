// Render thread repost of a flagged thread.
package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner        address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	srcBoardName         = "test-board"
	dstBoardName         = "test-board-2"
)

var repostThreadID boards.ID

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a board and a thread
	boardID := boards2.CreateBoard(cross, srcBoardName, false, false)
	threadID := boards2.CreateThread(cross, boardID, "Foo", "Body")

	// Repost thread into a different board
	dstBoardID := boards2.CreateBoard(cross, dstBoardName, false, false)
	repostThreadID = boards2.CreateRepost(cross, boardID, threadID, dstBoardID, "Bar", "Body2")

	// Flag original thread
	boards2.SetFlaggingThreshold(cross, boardID, 1)
	boards2.FlagThread(cross, boardID, threadID, "")
}

func main() {
	path := dstBoardName + "/" + repostThreadID.String()
	println(boards2.Render(path))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › [test\-board\-2](/r/gnoland/boards2/v1:test-board-2)
//
//
// > [!NOTE] 2 THINGS YOU NEED TO GET STARTED
// > To post and comment, **retrieve your Gno key** with [Adena](https://adena.app) or [gnokey](https://docs.gno.land/users/interact-with-gnokey). You need **at least 3,000 GNOT** in your account (balance requirement, not a transfer).
// ## Bar
//
// **[g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq)** `owner` · 2009-02-13 11:31pm UTC
// Body2
//
// > ⚠ Source post has been flagged as inappropriate
//
// ↳ [Flag](/r/gnoland/boards2/v1:test-board-2/1/flag) • [Repost](/r/gnoland/boards2/v1:test-board-2/1/repost) • [Comment](/r/gnoland/boards2/v1:test-board-2/1/reply) • [Edit](/r/gnoland/boards2/v1:test-board-2/1/edit) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteThread&boardID=2&threadID=1)
