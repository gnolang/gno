// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_set_permissions_01_filetest

package z_set_permissions_01_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/permissions"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	user address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	bid          = boards.ID(0)                               // Operate on realm instead of individual boards
)

var perms boards.Permissions

func init() {
	// Create a new permissions instance
	perms = permissions.New()
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	boards2.SetPermissions(cross(cur), bid, perms)
}

// Error:
// unauthorized, user g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj doesn't have the required permission
