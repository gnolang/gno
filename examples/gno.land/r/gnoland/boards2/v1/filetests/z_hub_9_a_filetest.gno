// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_9_a_filetest

// Test default thread comment values
package z_hub_9_a_filetest

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
// Comment
// false
// 0
// 0
// g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh
// 1234567890
// 0
