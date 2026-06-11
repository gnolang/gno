// PKGPATH: gno.land/r/gnoland/boards2/v1/hub/filetests/z_hub_8_b_filetest

// Test getting replies from a comment
package z_hub_8_b_filetest

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

	// Create replies
	boards2.CreateReply(cross(cur), boardID, threadID, commentID, "Reply 1")
	subCommentID := boards2.CreateReply(cross(cur), boardID, threadID, commentID, "Reply 2")

	// Add a sub-reply (it should not be included in the output)
	boards2.CreateReply(cross(cur), boardID, threadID, subCommentID, "Reply 3")

	// Get readonly comment
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	comment, _ = hub.GetComment(uint64(boardID), uint64(threadID), uint64(commentID))
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	replies := hub.GetReplies(comment.BoardID, comment.ThreadID, comment.ID, 0, comment.ReplyCount)

	for _, r := range replies {
		println(r.ID, r.Body)
	}
}

// Output:
// 3 Reply 1
// 4 Reply 2
