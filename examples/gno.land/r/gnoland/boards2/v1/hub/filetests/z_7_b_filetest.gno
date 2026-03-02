// Test getting comments from a thread
package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

var thread hub.Thread

func init() {
	testing.SetRealm(testing.NewUserRealm("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"))
	boardID := boards2.CreateBoard(cross, "test123", false, false)
	threadID := boards2.CreateThread(cross, boardID, "Title", "Body")
	commentID := boards2.CreateReply(cross, boardID, threadID, 0, "Comment 1")

	// Create another comment
	boards2.CreateReply(cross, boardID, threadID, 0, "Comment 2")

	// Add a reply (it should not be included in the output)
	boards2.CreateReply(cross, boardID, threadID, commentID, "Reply")

	// Get readonly thread
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	thread, _ = hub.GetThread(uint64(boardID), uint64(threadID))
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	comments := hub.GetComments(thread.BoardID, thread.ID, 0, thread.CommentCount)

	for _, c := range comments {
		println(c.ID, c.Body)
	}
}

// Output:
// 2 Comment 1
// 3 Comment 2
