// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_set_permissions_00_filetest

package z_set_permissions_00_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/permissions"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"
	bid           = boards.ID(0) // Operate on realm instead of individual boards
)

var perms boards.Permissions

func init() {
	// Create a new permissions instance without users
	perms = permissions.New()
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetPermissions(cross(cur), bid, perms)

	// Owner that setted new permissions is not a member of the new permissions
	println(boards2.IsMember(bid, owner))
}

// Output:
// false
