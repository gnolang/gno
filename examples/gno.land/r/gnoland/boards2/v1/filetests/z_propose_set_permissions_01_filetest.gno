// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_propose_set_permissions_01_filetest

// Test the normal flow for ProposeSetPermissions and GovDAO approval to set board permissions.

package z_propose_set_permissions_01_filetest

import (
	"chain/runtime/unsafe"
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/permissions"
	"gno.land/p/nt/testutils/v0"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gov/dao"
	daov3init "gno.land/r/gov/dao/v3/init"
)

const boardOwner address = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"

var (
	member   = unsafe.OriginCaller()
	newOwner = testutils.TestAddress("newOwner")
	bid      boards.ID
	perms    *permissions.Permissions
)

func init(cur realm) {
	daov3init.InitWithUsers(cross(cur), member)

	testing.SetRealm(testing.NewUserRealm(boardOwner))
	bid = boards2.CreateBoard(cross(cur), "test123", false, false)

	perms = permissions.New(
		permissions.UseSingleUserRole(),
		permissions.WithSuperRole("owner"),
	)
	perms.SetUserRoles(newOwner, "owner")

	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))

	pr := boards2.ProposeSetPermissions(cross(cur), bid, perms)
	dao.MustCreateProposal(cross(cur), pr)
}

func main(cur realm) {
	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))

	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(0)))
	println("executed =", dao.ExecuteProposal(cross(cur), dao.ProposalID(0)))

	println("new owner =", boards2.HasMemberRole(bid, newOwner, "owner"))
	println("board creator is owner =", boards2.HasMemberRole(bid, boardOwner, "owner"))

	// Realm permissions are untouched when a single board is targeted.
	println("realm untouched =", boards2.HasMemberRole(0, boardOwner, "owner"))
}

// Output:
// executed = true
// new owner = true
// board creator is owner = false
// realm untouched = true
