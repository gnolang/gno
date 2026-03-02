// Test getting replies from a comment without replies
package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

var comment hub.Comment

func init() {
	testing.SetRealm(testing.NewUserRealm("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"))
	boardID := boards2.CreateBoard(cross, "test123", false, false)
	threadID := boards2.CreateThread(cross, boardID, "Title", "Body")
	commentID := boards2.CreateReply(cross, boardID, threadID, 0, "Comment")

	// Get readonly comment
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	comment, _ = hub.GetComment(uint64(boardID), uint64(threadID), uint64(commentID))
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	replies := hub.GetReplies(comment.BoardID, comment.ThreadID, comment.ID, 0, comment.ReplyCount)
	println(len(replies))
}

// Output:
// 0
