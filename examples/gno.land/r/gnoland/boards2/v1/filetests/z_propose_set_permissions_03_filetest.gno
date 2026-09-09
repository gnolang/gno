// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_propose_set_permissions_03_filetest

// Test the failure for ProposeSetPermissions and GovDAO approval where the
// realm is locked between proposal creation and execution.

package z_propose_set_permissions_03_filetest

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

const realmOwner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

var (
	member   = unsafe.OriginCaller()
	newOwner = testutils.TestAddress("newOwner")
	perms    boards.Permissions
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
	testing.SetRealm(testing.NewUserRealm(realmOwner))
	boards2.LockRealm(cross(cur), true)

	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))

	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(0)))
	println("executed =", dao.ExecuteOrRejectProposal(cross(cur), dao.ProposalID(0)))
	println("new owner =", boards2.HasMemberRole(0, newOwner, "owner"))
}

// Output:
// executed = false
// new owner = false
