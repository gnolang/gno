// PKGPATH: gno.land/r/gnoland/boards2/v1/hub/filetests/z_6_b_filetest

// Test getting flags from an unflagged comment
package z_6_b_filetest

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

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boardID = boards2.CreateBoard(cross(cur), "test123", false, false)
	threadID = boards2.CreateThread(cross(cur), boardID, "Title", "Body")
	commentID = boards2.CreateReply(cross(cur), boardID, threadID, 0, "Comment")
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	flags := hub.GetFlags(uint64(boardID), uint64(threadID), uint64(commentID), 0, 1)
	println(len(flags))
}

// Output:
// 0
