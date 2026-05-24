// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_accept_invite_00_filetest

package z_accept_invite_00_filetest

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

	testing.SetRealm(testing.NewUserRealm(user))
	boards2.RequestInvite(cross(cur), bid)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.AcceptInvite(cross(cur), bid, user)

	println(boards2.IsMember(bid, user))
	println()
	println(boards2.Render("test123/invites"))
}

// Output:
// true
//
// # test123 Invite Requests
// ### Board has no invite requests
