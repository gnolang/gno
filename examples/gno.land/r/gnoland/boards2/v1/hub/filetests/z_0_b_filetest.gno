// PKGPATH: gno.land/r/gnoland/boards2/v1/hub/filetests/z_0_b_filetest

// Test non default board values
package z_0_b_filetest

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

	// Invite member
	boards2.InviteMember(cross(cur), boardID, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", "admin")

	// Rename board
	boards2.RenameBoard(cross(cur), "test123", "foo123")

	// Create a thread
	boards2.CreateThread(cross(cur), boardID, "Title", "Body")

	// Freeze board
	boards2.FreezeBoard(cross(cur), boardID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	board, found := hub.GetBoard(uint64(boardID))
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
// foo123
// slice[("test123" string)]
// true
// 1
// 2
// g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh
// 1234567890
// 1234567890
