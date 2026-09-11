// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_3_a_filetest

// Test getting threads from an empty board
package z_hub_3_a_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards/v0"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

var boardID boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"))
	boardID = boards2.CreateBoard(cross(cur), "test123", false, false)
}

func main(cur realm) {
	boards := boards2.GetThreads(uint64(boardID), 0, boards2.BoardCount())
	println(len(boards))
}

// Output:
// 0
