// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_5_b_filetest

// Test getting reposts of a thread
package z_hub_5_b_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

var thread boards2.Thread

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boardID := boards2.CreateBoard(cross(cur), "aaa123", false, false)
	threadID := boards2.CreateThread(cross(cur), boardID, "Title", "Body")

	// Create first repost
	dstBoardID := boards2.CreateBoard(cross(cur), "bbb123", false, false)
	boards2.CreateRepost(cross(cur), boardID, threadID, dstBoardID, "First", "Body")

	// Create second repost
	dstBoardID = boards2.CreateBoard(cross(cur), "ccc123", false, false)
	boards2.CreateRepost(cross(cur), boardID, threadID, dstBoardID, "Second", "Body")

	// Get readonly thread
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	thread, _ = boards2.GetThread(uint64(boardID), uint64(threadID))
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	reposts := boards2.GetReposts(thread.BoardID, thread.ID, 0, thread.RepostCount)

	for _, t := range reposts {
		println(t.ID, t.Title)
	}
}

// Output:
// 1 First
// 1 Second
