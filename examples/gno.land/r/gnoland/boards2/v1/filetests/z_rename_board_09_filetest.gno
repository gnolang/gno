// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_rename_board_09_filetest

package z_rename_board_09_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	name          = "foo123"
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.CreateBoard(cross(cur), name, false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	boards2.RenameBoard(cross(cur), name, "barbaz")
}

// Error:
// unauthorized, user g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj doesn't have the required permission
