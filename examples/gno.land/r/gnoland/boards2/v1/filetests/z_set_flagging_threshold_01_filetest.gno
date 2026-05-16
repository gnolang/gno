// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_set_flagging_threshold_01_filetest

package z_set_flagging_threshold_01_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address   = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	bid   boards.ID = 404
)

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetFlaggingThreshold(cross2(cur), bid, 1)
}

// Error:
// board does not exist with ID: 404
