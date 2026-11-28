// PKGPATH: gno.land/r/gnoland/boards2/v0/filetests/z_propose_set_permissions_02_filetest

// Test "board not found" for ProposeSetPermissions and GovDAO approval.

package z_propose_set_permissions_02_filetest

import (
	"chain/runtime/unsafe"
	"testing"

	"gno.land/p/gnoland/boards/exts/permissions/v0"
	"gno.land/p/gnoland/boards/v0"

	boards2 "gno.land/r/gnoland/boards2/v0"
	daov3init "gno.land/r/gov/dao/init/v0"
)

var member = unsafe.OriginCaller()

func init(cur realm) {
	daov3init.InitWithUsers(cross(cur), member)
}

func main(cur realm) {
	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))

	boards2.ProposeSetPermissions(cross(cur), boards.ID(99), permissions.New())
}

// Error:
// board not found: 99
