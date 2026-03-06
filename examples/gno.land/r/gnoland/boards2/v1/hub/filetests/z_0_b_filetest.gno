// Test non default board values
package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

var boardID boards.ID

func init() {
	testing.SetRealm(testing.NewUserRealm("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"))
	boardID = boards2.CreateBoard(cross, "test123", false, false)

	// Invite member
	boards2.InviteMember(cross, boardID, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", "admin")

	// Rename board
	boards2.RenameBoard(cross, "test123", "foo123")

	// Create a thread
	boards2.CreateThread(cross, boardID, "Title", "Body")

	// Freeze board
	boards2.FreezeBoard(cross, boardID)
}

func main() {
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
// g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq
// 1234567890
// 1234567890
