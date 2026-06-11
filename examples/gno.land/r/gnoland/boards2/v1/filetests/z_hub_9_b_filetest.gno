// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_9_b_filetest

// Test non default thread comment values
package z_hub_9_b_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

var (
	boardID   boards.ID
	threadID  boards.ID
	commentID boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boardID = boards2.CreateBoard(cross(cur), "origin123", false, false)
	threadID = boards2.CreateThread(cross(cur), boardID, "Title", "Body")
	commentID = boards2.CreateReply(cross(cur), boardID, threadID, 0, "Comment")

	// Edit comment
	boards2.EditReply(cross(cur), boardID, threadID, commentID, "Test comment")

	// Create a comment reply
	boards2.CreateReply(cross(cur), boardID, threadID, commentID, "Reply")

	// Flag comment
	boards2.FlagReply(cross(cur), boardID, threadID, commentID, "Reason")
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	comment, found := boards2.GetComment(uint64(boardID), uint64(threadID), uint64(commentID))
	if !found {
		return
	}

	println(comment.ID)
	println(comment.BoardID)
	println(comment.ThreadID)
	println(comment.ParentID)
	println(comment.Body)
	println(comment.Hidden)
	println(comment.ReplyCount)
	println(comment.FlagCount)
	println(comment.Creator)
	println(comment.CreatedAt)
	println(comment.UpdatedAt)
}

// Output:
// 2
// 1
// 1
// 1
// Test comment
// true
// 1
// 1
// g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh
// 1234567890
// 1234567890
