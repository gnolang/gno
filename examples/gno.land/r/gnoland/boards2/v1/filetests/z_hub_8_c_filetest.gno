// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_8_c_filetest

// Test getting replies from a nested reply
package z_hub_8_c_filetest

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
	testing.SetRealm(testing.NewUserRealm("g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"))
	boardID = boards2.CreateBoard(cross(cur), "test123", false, false)
	threadID = boards2.CreateThread(cross(cur), boardID, "Title", "Body")
	commentID := boards2.CreateReply(cross(cur), boardID, threadID, 0, "Comment")
	replyID = boards2.CreateReply(cross(cur), boardID, threadID, commentID, "Reply")

	// Create sub-replies of the nested reply
	boards2.CreateReply(cross(cur), boardID, threadID, replyID, "Sub-reply 1")
	boards2.CreateReply(cross(cur), boardID, threadID, replyID, "Sub-reply 2")
}

func main(cur realm) {
	replies := boards2.GetReplies(uint64(boardID), uint64(threadID), uint64(replyID), 0, 10)

	for _, r := range replies {
		println(r.ID(), r.Body())
	}
}

// Output:
// 4 Sub-reply 1
// 5 Sub-reply 2
