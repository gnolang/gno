// PKGPATH: gno.land/r/gnoland/boards2/v1/hub/filetests/z_hub_8_a_filetest

// Test getting replies from a comment without replies
package z_hub_8_a_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

var comment hub.Comment

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boardID := boards2.CreateBoard(cross(cur), "test123", false, false)
	threadID := boards2.CreateThread(cross(cur), boardID, "Title", "Body")
	commentID := boards2.CreateReply(cross(cur), boardID, threadID, 0, "Comment")

	// Get readonly comment
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	comment, _ = hub.GetComment(uint64(boardID), uint64(threadID), uint64(commentID))
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	replies := hub.GetReplies(comment.BoardID, comment.ThreadID, comment.ID, 0, comment.ReplyCount)
	println(len(replies))
}

// Output:
// 0
