// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_get_board_id_from_name_01_filetest

package z_get_board_id_from_name_01_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	bid, found := boards2.GetBoardIDFromName(cross(cur), "foobar")

	println(found)
	println(bid == 0)
}

// Output:
// false
// true
