// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_board_03_filetest

package z_create_board_03_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateBoard(cross(cur), "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", false, false)
}

// Error:
// addresses are not allowed as board name
