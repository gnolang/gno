// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_9_c_filetest

// Test getting a nested reply by ID
package z_hub_9_c_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

var (
	boardID  boards.ID
	threadID boards.ID
	replyID  boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boardID = boards2.CreateBoard(cross(cur), "test123", false, false)
	threadID = boards2.CreateThread(cross(cur), boardID, "Title", "Body")
	commentID := boards2.CreateReply(cross(cur), boardID, threadID, 0, "Comment")

	// Reply to the comment, so the reply is not a top level comment
	replyID = boards2.CreateReply(cross(cur), boardID, threadID, commentID, "Reply")
}

func main(cur realm) {
	reply, found := boards2.GetComment(uint64(boardID), uint64(threadID), uint64(replyID))
	println(found)
	println(reply.ID())
	println(reply.ParentID())
	println(reply.Body())
}

// Output:
// true
// 3
// 2
// Reply
