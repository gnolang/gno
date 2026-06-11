// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_get_board_id_from_name_00_filetest

package z_get_board_id_from_name_00_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	name          = "test123"
)

var bid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), name, false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	bid2, found := boards2.GetBoardIDFromName(cross(cur), name)

	println(found)
	println(bid2 == bid)
}

// Output:
// true
// true
