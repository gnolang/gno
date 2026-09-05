// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ban_01_filetest

package z_ban_01_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"
	user  address = "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"
)

var bid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test123", false, false)

	// Invite the user as a moderator
	boards2.InviteMember(cross(cur), bid, user, boards2.RoleModerator)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.Ban(cross(cur), bid, user, boards2.BanDay, "Reason")
}

// Error:
// owner, admin and moderator banning is not allowed
