// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_board_10_filetest

package z_create_board_10_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Should fail because board name has a space which is not allowed
	boards2.CreateBoard(cross(cur), "test 123", false, false)
}

// Error:
// board name must start with a letter and have letters, numbers, "-" and "_"
