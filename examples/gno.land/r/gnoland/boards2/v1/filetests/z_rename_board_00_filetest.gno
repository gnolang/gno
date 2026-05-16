// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_rename_board_00_filetest

package z_rename_board_00_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	name            = "foo123"
	newName         = "bar123"
)

var bid boards.ID // Operate on board DAO

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross2(cur), name, false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.RenameBoard(cross2(cur), name, newName)

	// Ensure board is renamed by the default board owner
	bid2, _ := boards2.GetBoardIDFromName(cross2(cur), newName)
	println("IDs match =", bid == bid2)
}

// Output:
// IDs match = true
