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
	testing.SetRealm(testing.NewUserRealm("g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"))
	boardID = boards2.CreateBoard(cross(cur), "test123", false, false)
}

func main(cur realm) {
	board, found := boards2.GetBoard(uint64(boardID))
	if !found {
		return
	}

	println(board.ID())
	println(board.Name())
	println(board.Aliases())
	println(board.Readonly())
	println(board.ThreadCount())
	println(board.MemberCount())
	println(board.Creator())
	println(board.CreatedAt())
	println(board.UpdatedAt())
}

// Output:
// 2
// test123
// (nil []string)
// false
// 0
// 1
// g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf
// 1234567890
// 0
