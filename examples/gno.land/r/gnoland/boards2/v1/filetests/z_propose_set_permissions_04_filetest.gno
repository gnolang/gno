// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_propose_set_permissions_04_filetest

// Test the following ownership flow for ProposeSetPermissions and GovDAO approval:
// The permissions installed by a proposal stay mutable afterwards. They are
// owned by the realm that allocated them, so boards2 mutates them under that
// realm's authority (borrow rule #2) rather than its own.

package z_propose_set_permissions_04_filetest

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

var (
	member   = unsafe.OriginCaller()
	newOwner = testutils.TestAddress("newOwner")
	invitee  = testutils.TestAddress("invitee")
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
	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewUserRealm(member))

	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(0)))
	println("executed =", dao.ExecuteProposal(cross(cur), dao.ProposalID(0)))

	testing.SetOriginCaller(newOwner)
	testing.SetRealm(testing.NewUserRealm(newOwner))
	boards2.InviteMember(cross(cur), 0, invitee, "owner")

	println("invitee is member =", boards2.IsMember(0, invitee))
}

// Output:
// executed = true
// invitee is member = true
