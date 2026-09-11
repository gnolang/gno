// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_propose_set_permissions_00_filetest

// Test the normal flow for ProposeSetPermissions and GovDAO approval to set realm permissions.

package z_propose_set_permissions_00_filetest

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

// member submits the proposal: only GovDAO members can. It holds no boards2
// role, so it could not call SetPermissions.
var (
	member   = unsafe.OriginCaller()
	newOwner = testutils.TestAddress("newOwner")
	perms    *permissions.Permissions
)

func init(cur realm) {
	daov3init.InitWithUsers(cross(cur), member)

	perms = permissions.New(
		permissions.UseSingleUserRole(),
		permissions.WithSuperRole("owner"),
	)
	perms.SetUserRoles(newOwner, "owner")

	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))

	pr := boards2.ProposeSetPermissions(cross(cur), boards.ID(0), perms)
	dao.MustCreateProposal(cross(cur), pr)
}

func main(cur realm) {
	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))

	println("new owner before =", boards2.HasMemberRole(0, newOwner, "owner"))

	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(0)))
	println("executed =", dao.ExecuteProposal(cross(cur), dao.ProposalID(0)))

	println("new owner after =", boards2.HasMemberRole(0, newOwner, "owner"))
	println("proposer is owner =", boards2.HasMemberRole(0, member, "owner"))
}

// Output:
// new owner before = false
// executed = true
// new owner after = true
// proposer is owner = false
