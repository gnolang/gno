// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_set_permissions_02_filetest

package z_set_permissions_02_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/permissions"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var (
	perms boards.Permissions
	bid   boards.ID
)

func init(cur realm) {
	// Create a new permissions instance without users
	perms = permissions.New()

	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "foobar", false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetPermissions(cross(cur), bid, perms)

	// Owner that setted new board permissions is not a member of the new permissions
	println(boards2.IsMember(bid, owner))
}

// Output:
// false
