// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_rename_board_02_filetest

package z_rename_board_02_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"
	name          = "foo123"
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.CreateBoard(cross(cur), name, false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.RenameBoard(cross(cur), name, name)
}

// Error:
// board already exists
