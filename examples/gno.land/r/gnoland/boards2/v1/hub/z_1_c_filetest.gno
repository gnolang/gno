// Test non default reposted thread values
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
	srcBoardID := boards2.CreateBoard(cross, "origin123", false, false)
	srcThreadID := boards2.CreateThread(cross, srcBoardID, "Title", "Body")
	boardID = boards2.CreateBoard(cross, "test123", false, false)
	threadID = boards2.CreateRepost(cross, srcBoardID, srcThreadID, boardID, "Title", "Body")

	// Edit repost
	boards2.EditThread(cross, boardID, threadID, "Foo", "Bar")

	// Add a comment to the repost
	boards2.CreateReply(cross, boardID, threadID, 0, "Comment")

	// Freeze repost
	boards2.FreezeThread(cross, boardID, threadID)

	// Flag repost
	boards2.FlagThread(cross, boardID, threadID, "Reason")
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	thread, found := hub.GetThread(uint64(boardID), uint64(threadID))
	if !found {
		return
	}

	println(thread.ID)
	println(thread.OriginalBoardID)
	println(thread.BoardID)
	println(thread.Title)
	println(thread.Body)
	println(thread.Hidden)
	println(thread.Readonly)
	println(thread.CommentCount)
	println(thread.RepostCount) // Reposts can't be reposted
	println(thread.FlagCount)
	println(thread.Creator)
	println(thread.CreatedAt)
	println(thread.UpdatedAt)
}

// Output:
// 1
// 1
// 2
// Foo
// Bar
// true
// true
// 1
// 0
// 1
// g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq
// 1234567890
// 1234567890
