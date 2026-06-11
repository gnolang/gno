// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_7_b_filetest

// Test getting comments from a thread
package z_hub_7_b_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

var thread boards2.Thread

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boardID := boards2.CreateBoard(cross(cur), "test123", false, false)
	threadID := boards2.CreateThread(cross(cur), boardID, "Title", "Body")
	commentID := boards2.CreateReply(cross(cur), boardID, threadID, 0, "Comment 1")

	// Create another comment
	boards2.CreateReply(cross(cur), boardID, threadID, 0, "Comment 2")

	// Add a reply (it should not be included in the output)
	boards2.CreateReply(cross(cur), boardID, threadID, commentID, "Reply")

	// Get readonly thread
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	thread, _ = boards2.GetThread(uint64(boardID), uint64(threadID))
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	comments := boards2.GetComments(thread.BoardID, thread.ID, 0, thread.CommentCount)

	for _, c := range comments {
		println(c.ID, c.Body)
	}
}

// Output:
// 2 Comment 1
// 3 Comment 2
