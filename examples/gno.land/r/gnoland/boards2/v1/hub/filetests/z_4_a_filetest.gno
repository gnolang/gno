// PKGPATH: gno.land/r/gnoland/boards2/v1/hub/filetests/z_4_a_filetest

// Test getting board members
package z_4_a_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

var board hub.Board

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boardID := boards2.CreateBoard(cross(cur), "test123", false, false)

	// Invite board members
	boards2.InviteMember(cross(cur), boardID, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "owner")
	boards2.InviteMember(cross(cur), boardID, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", "admin")

	// Get readonly board
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	board, _ = hub.GetBoard(uint64(boardID))
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	members := hub.GetMembers(board.ID, 0, board.MemberCount)

	for _, m := range members {
		println(m.Address, m.Roles)
	}
}

// Output:
// g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 slice[("owner" gno.land/p/gnoland/boards.Role)]
// g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh slice[("owner" gno.land/p/gnoland/boards.Role)]
// g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj slice[("admin" gno.land/p/gnoland/boards.Role)]
