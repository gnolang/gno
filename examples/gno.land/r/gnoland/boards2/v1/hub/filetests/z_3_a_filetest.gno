// PKGPATH: gno.land/r/gnoland/boards2/v1/hub/filetests/z_3_a_filetest

// Test getting threads from an empty board
package z_3_a_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

var boardID boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boardID = boards2.CreateBoard(cross(cur), "test123", false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	boards := hub.GetThreads(uint64(boardID), 0, boards2.BoardCount())
	println(len(boards))
}

// Output:
// 0
