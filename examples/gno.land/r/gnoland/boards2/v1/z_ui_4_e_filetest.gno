// Render thread repost of a deleted thread.
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

	// Remove the original thread
	boards2.DeleteThread(cross, boardID, threadID)
}

func main() {
	path := dstBoardName + "/" + repostThreadID.String()
	println(boards2.Render(path))
}

// Output:
// # Boards › test-board-2
// ## Bar
//
// > *⚠ Source post has been deleted*
// >
// Body2
//
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC
// [Flag](/r/gnoland/boards2/v1:test-board-2/1/flag) • [Repost](/r/gnoland/boards2/v1:test-board-2/1/repost) • [Comment](/r/gnoland/boards2/v1:test-board-2/1/reply) • [Edit](/r/gnoland/boards2/v1:test-board-2/1/edit) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteThread&boardID=2&threadID=1)
