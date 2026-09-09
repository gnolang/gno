// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_propose_set_permissions_02_filetest

// Test "board not found" for ProposeSetPermissions and GovDAO approval.

package z_propose_set_permissions_02_filetest

import (
	"chain/runtime/unsafe"
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/permissions"

	boards2 "gno.land/r/gnoland/boards2/v1"
	daov3init "gno.land/r/gov/dao/v3/init"
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
