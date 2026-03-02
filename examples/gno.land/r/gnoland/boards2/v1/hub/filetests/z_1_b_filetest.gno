// Test non default thread values
package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

var (
	boardID  boards.ID
	threadID boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"))
	boardID = boards2.CreateBoard(cross, "test123", false, false)
	threadID = boards2.CreateThread(cross, boardID, "Title", "Body")

	// Create another board and repost thread
	dstBoardID := boards2.CreateBoard(cross, "destination123", false, false)
	boards2.CreateRepost(cross, boardID, threadID, dstBoardID, "Title", "Body")

	// Edit thread
	boards2.EditThread(cross, boardID, threadID, "Foo", "Bar")

	// Add a comment to the thread
	boards2.CreateReply(cross, boardID, threadID, 0, "Comment")

	// Freeze thread
	boards2.FreezeThread(cross, boardID, threadID)

	// Flag thread
	boards2.FlagThread(cross, boardID, threadID, "Reason")
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	thread, found := hub.GetThread(uint64(boardID), uint64(threadID))
	if !found {
		return
	}

	println(thread.ID)
	println(thread.OriginalBoardID) // Only reposts have an original board ID
	println(thread.BoardID)
	println(thread.Title)
	println(thread.Body)
	println(thread.Hidden)
	println(thread.Readonly)
	println(thread.CommentCount)
	println(thread.RepostCount)
	println(thread.FlagCount)
	println(thread.Creator)
	println(thread.CreatedAt)
	println(thread.UpdatedAt)
}

// Output:
// 1
// 0
// 1
// Foo
// Bar
// true
// true
// 1
// 1
// 1
// g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq
// 1234567890
// 1234567890
