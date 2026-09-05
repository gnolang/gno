// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_rename_board_03_filetest

package z_rename_board_03_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.RenameBoard(cross(cur), "unexisting", "foo")
}

// Error:
// board does not exist with name: unexisting
