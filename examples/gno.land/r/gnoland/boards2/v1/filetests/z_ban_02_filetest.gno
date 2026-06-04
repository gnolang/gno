// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ban_02_filetest

package z_ban_02_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	user  address = "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"
)

var bid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test123", false, false)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Try to ban a user without banning permissions
	boards2.Ban(cross(cur), bid, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", boards2.BanDay, "Reason")
}

// Error:
// unauthorized, user g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5 doesn't have the required permission
