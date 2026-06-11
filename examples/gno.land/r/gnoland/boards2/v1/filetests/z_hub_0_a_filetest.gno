// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_0_a_filetest

// Test default board values
package z_hub_0_a_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

var boardID boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boardID = boards2.CreateBoard(cross(cur), "test123", false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	board, found := boards2.GetBoard(uint64(boardID))
	if !found {
		return
	}

	println(board.ID)
	println(board.Name)
	println(board.Aliases)
	println(board.Readonly)
	println(board.ThreadCount)
	println(board.MemberCount)
	println(board.Creator)
	println(board.CreatedAt)
	println(board.UpdatedAt)
}

// Output:
// 1
// test123
// (nil []string)
// false
// 0
// 1
// g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh
// 1234567890
// 0
