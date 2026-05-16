// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_rename_board_05_filetest

package z_rename_board_05_filetest

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	member  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	name            = "foo123"
	newName         = "barbaz123"
)

var bid boards.ID // Operate on board DAO

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	bid = boards2.CreateBoard(cross2(cur), name, false, false)
	boards2.InviteMember(cross2(cur), bid, member, boards2.RoleOwner)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(member))

	boards2.RenameBoard(cross2(cur), name, newName)

	// Ensure board is renamed by another board owner
	bid2, _ := boards2.GetBoardIDFromName(cross2(cur), newName)
	println("IDs match =", bid == bid2)
}

// Output:
// IDs match = true
