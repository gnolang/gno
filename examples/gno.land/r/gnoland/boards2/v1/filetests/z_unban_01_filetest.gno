// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_unban_01_filetest

package z_unban_01_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var bid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross2(cur), "test123", false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.Unban(cross2(cur), bid, "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5", "")
}

// Error:
// user is not banned
