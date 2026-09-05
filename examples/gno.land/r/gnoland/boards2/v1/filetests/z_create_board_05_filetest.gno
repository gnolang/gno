// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_board_05_filetest

package z_create_board_05_filetest

import (
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"

var name = strings.Repeat("X", boards2.MaxBoardNameLength+1)

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateBoard(cross(cur), name, false, false)
}

// Error:
// board name is too long, maximum allowed is 50 characters
