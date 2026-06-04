// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_unban_02_filetest

package z_unban_02_filetest

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
	boards2.Ban(cross(cur), bid, user, boards2.BanDay, "Unpolite behavior")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"))

	// Try to unban without unbanning permissions
	boards2.Unban(cross(cur), bid, user, "")
}

// Error:
// unauthorized, user g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj doesn't have the required permission
