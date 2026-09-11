// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_propose_create_board_01_filetest

// Test "board already exists" for ProposeCreateBoard and GovDAO approval.

package z_propose_create_board_01_filetest

import (
	"chain/runtime/unsafe"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	daov3init "gno.land/r/gov/dao/v3/init"
)

const (
	owner     address = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"
	boardName         = "test123"
)

var member address = unsafe.OriginCaller()

func init(cur realm) {
	daov3init.InitWithUsers(cross(cur), member)

	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.CreateBoard(cross(cur), boardName, false, false)
}

func main(cur realm) {
	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))

	boards2.ProposeCreateBoard(cross(cur), boardName, true, false, member)
}

// Error:
// board already exists
