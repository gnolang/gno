// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_1_c_filetest

// Test non default reposted thread values
package z_hub_1_c_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

var (
	boardID  boards.ID
	threadID boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	srcBoardID := boards2.CreateBoard(cross(cur), "origin123", false, false)

	// Create two threads where the second is the one to repost
	boards2.CreateThread(cross(cur), srcBoardID, "Title1", "Body1")
	srcThreadID := boards2.CreateThread(cross(cur), srcBoardID, "Title2", "Body2") // ID = 2

	// Create repost
	boardID = boards2.CreateBoard(cross(cur), "test123", false, false)
	threadID = boards2.CreateRepost(cross(cur), srcBoardID, srcThreadID, boardID, "Title", "Body")

	// Edit repost
	boards2.EditThread(cross(cur), boardID, threadID, "Foo", "Bar")

	// Add a comment to the repost
	boards2.CreateReply(cross(cur), boardID, threadID, 0, "Comment")

	// Freeze repost
	boards2.FreezeThread(cross(cur), boardID, threadID)

	// Flag repost
	boards2.FlagThread(cross(cur), boardID, threadID, "Reason")
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	thread, found := boards2.GetThread(uint64(boardID), uint64(threadID))
	if !found {
		return
	}

	println(thread.ID)
	println(thread.OriginalBoardID)
	println(thread.OriginalThreadID)
	println(thread.BoardID)
	println(thread.Title)
	println(thread.Body)
	println(thread.Hidden)
	println(thread.Readonly)
	println(thread.CommentCount)
	println(thread.RepostCount) // Reposts can't be reposted, so count must be 0
	println(thread.FlagCount)
	println(thread.Creator)
	println(thread.CreatedAt)
	println(thread.UpdatedAt)
}

// Output:
// 1
// 1
// 2
// 2
// Foo
// Bar
// true
// true
// 1
// 0
// 1
// g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh
// 1234567890
// 1234567890
