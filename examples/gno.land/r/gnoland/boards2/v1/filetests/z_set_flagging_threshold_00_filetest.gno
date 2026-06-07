// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_set_flagging_threshold_00_filetest

package z_set_flagging_threshold_00_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var bid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test-board", false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetFlaggingThreshold(cross(cur), bid, 4)

	// Ensure that flagging threshold changed
	println(boards2.GetFlaggingThreshold(bid))
}

// Output:
// 4
