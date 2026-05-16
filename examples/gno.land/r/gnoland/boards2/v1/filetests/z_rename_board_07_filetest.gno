// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_rename_board_07_filetest

package z_rename_board_07_filetest

import (
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	name          = "foo123"
)

var newName string

func init(cur realm) {
	newName = strings.Repeat("A", boards2.MaxBoardNameLength+1)

	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.CreateBoard(cross2(cur), name, false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.RenameBoard(cross2(cur), name, newName)
}

// Error:
// board name is too long, maximum allowed is 50 characters
