// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_1_a_filetest

// Test default thread values
package z_hub_1_a_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

var (
	boardID  boards.ID
	threadID boards.ID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boardID = boards2.CreateBoard(cross(cur), "origin123", false, false)
	threadID = boards2.CreateThread(cross(cur), boardID, "Title", "Body")
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	thread, found := boards2.GetThread(uint64(boardID), uint64(threadID))
	if !found {
		return
	}

	println(thread.ID)
	println(thread.OriginalBoardID)
	println(thread.OriginalThreadID)
	println(thread.BoardID)
	println(thread.Title)
	println(thread.Body)
	println(thread.Hidden)
	println(thread.Readonly)
	println(thread.CommentCount)
	println(thread.RepostCount)
	println(thread.FlagCount)
	println(thread.Creator)
	println(thread.CreatedAt)
	println(thread.UpdatedAt)
}

// Output:
// 1
// 0
// 0
// 1
// Title
// Body
// false
// false
// 0
// 0
// 0
// g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh
// 1234567890
// 0
