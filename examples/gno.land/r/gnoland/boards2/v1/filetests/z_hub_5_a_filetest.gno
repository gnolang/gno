// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_5_a_filetest

// Test getting reposts from a thread without reposts
package z_hub_5_a_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

var thread boards2.Thread

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boardID := boards2.CreateBoard(cross(cur), "aaa123", false, false)
	threadID := boards2.CreateThread(cross(cur), boardID, "Title", "Body")

	// Get readonly thread
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	thread, _ = boards2.GetThread(uint64(boardID), uint64(threadID))
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	reposts := boards2.GetReposts(thread.BoardID, thread.ID, 0, thread.RepostCount)
	println(len(reposts))
}

// Output:
// 0
