// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_get_board_id_from_name_01_filetest

package z_get_board_id_from_name_01_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	bid, found := boards2.GetBoardIDFromName(cross(cur), "foobar")

	println(found)
	println(bid == 0)
}

// Output:
// false
// true
