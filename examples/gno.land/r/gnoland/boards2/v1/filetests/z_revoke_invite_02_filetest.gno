// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_revoke_invite_02_filetest

package z_revoke_invite_02_filetest

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
	bid = boards2.CreateBoard(cross2(cur), "test123", false, false)

	testing.SetRealm(testing.NewUserRealm(user))
	boards2.RequestInvite(cross2(cur), bid)
}

func main(cur realm) {
	// Caller is not a member and has no permission to revoke invites
	testing.SetRealm(testing.NewUserRealm(user))

	boards2.RevokeInvite(cross2(cur), bid, user)
}

// Error:
// unauthorized, user g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5 doesn't have the required permission
