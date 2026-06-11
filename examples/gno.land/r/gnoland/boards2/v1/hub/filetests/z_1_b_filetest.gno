// PKGPATH: gno.land/r/gnoland/boards2/v1/hub/filetests/z_1_b_filetest

// Test non default thread values
package z_1_b_filetest

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

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boardID = boards2.CreateBoard(cross(cur), "test123", false, false)
	threadID = boards2.CreateThread(cross(cur), boardID, "Title", "Body")

	// Create another board and repost thread
	dstBoardID := boards2.CreateBoard(cross(cur), "destination123", false, false)
	boards2.CreateRepost(cross(cur), boardID, threadID, dstBoardID, "Title", "Body")

	// Edit thread
	boards2.EditThread(cross(cur), boardID, threadID, "Foo", "Bar")

	// Add a comment to the thread
	boards2.CreateReply(cross(cur), boardID, threadID, 0, "Comment")

	// Freeze thread
	boards2.FreezeThread(cross(cur), boardID, threadID)

	// Flag thread
	boards2.FlagThread(cross(cur), boardID, threadID, "Reason")
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	thread, found := hub.GetThread(uint64(boardID), uint64(threadID))
	if !found {
		return
	}

	println(thread.ID)
	println(thread.OriginalBoardID)  // Only reposts have an original board ID
	println(thread.OriginalThreadID) // Only reposts have an original thread ID
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
// 0
// 1
// Foo
// Bar
// true
// true
// 1
// 1
// 1
// g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh
// 1234567890
// 1234567890
