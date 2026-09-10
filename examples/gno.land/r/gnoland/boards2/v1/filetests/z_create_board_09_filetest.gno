// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_board_09_filetest

package z_create_board_09_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.CreateBoard(cross(cur), "TEST123", false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Should fail because board name already exists with a different casing
	boards2.CreateBoard(cross(cur), "test123", false, false)
}

// Error:
// board already exists
