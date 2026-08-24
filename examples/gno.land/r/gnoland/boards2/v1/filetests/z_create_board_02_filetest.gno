// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_create_board_02_filetest

package z_create_board_02_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	boardName         = "test123"
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.CreateBoard(cross(cur), boardName, false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateBoard(cross(cur), boardName, false, false)
}

// Error:
// board already exists
