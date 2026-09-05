// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_freeze_board_01_filetest

package z_freeze_board_01_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"

var bid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test123", false, false)
	boards2.FreezeBoard(cross(cur), bid)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.FreezeBoard(cross(cur), bid)
}

// Error:
// board is frozen
