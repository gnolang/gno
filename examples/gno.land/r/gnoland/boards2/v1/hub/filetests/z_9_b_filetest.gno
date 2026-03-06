// Test non default thread comment values
package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

var (
	boardID   boards.ID
	threadID  boards.ID
	commentID boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"))
	boardID = boards2.CreateBoard(cross, "origin123", false, false)
	threadID = boards2.CreateThread(cross, boardID, "Title", "Body")
	commentID = boards2.CreateReply(cross, boardID, threadID, 0, "Comment")

	// Edit comment
	boards2.EditReply(cross, boardID, threadID, commentID, "Test comment")

	// Create a comment reply
	boards2.CreateReply(cross, boardID, threadID, commentID, "Reply")

	// Flag comment
	boards2.FlagReply(cross, boardID, threadID, commentID, "Reason")
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	comment, found := hub.GetComment(uint64(boardID), uint64(threadID), uint64(commentID))
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
// g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq
// 1234567890
// 1234567890
