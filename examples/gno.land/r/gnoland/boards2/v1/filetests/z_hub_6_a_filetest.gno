// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_6_a_filetest

// Test getting flags from an unflagged thread
package z_hub_6_a_filetest

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
	boardID = boards2.CreateBoard(cross(cur), "test123", false, false)
	threadID = boards2.CreateThread(cross(cur), boardID, "Title", "Body")
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	flags := boards2.GetFlags(uint64(boardID), uint64(threadID), 0, 0, 1)
	println(len(flags))
}

// Output:
// 0
