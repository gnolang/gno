// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_is_banned_00_filetest

package z_is_banned_00_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var bid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test123", false, false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	println(boards2.IsBanned(bid, "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"))
}

// Output:
// false
