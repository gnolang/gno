// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_propose_set_permissions_04_filetest

// Test that the realm permissions set by ProposeSetPermissions keep the
// validators of the permissions they replace.

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
	member = unsafe.OriginCaller()
	admin  = testutils.TestAddress("admin")
	perms  *permissions.Permissions
)

func init(cur realm) {
	daov3init.InitWithUsers(cross(cur), member)

	perms = permissions.New(
		permissions.UseSingleUserRole(),
		permissions.WithSuperRole(boards2.RoleOwner),
	)
	perms.AddRole(boards2.RoleAdmin, boards2.PermissionBoardCreate)
	perms.SetUserRoles(admin, boards2.RoleAdmin)

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

	testing.SetOriginCaller(admin)
	testing.SetRealm(testing.NewUserRealm(admin))
	boards2.CreateBoard(cross(cur), "test123", false, true)
}

// Output:
// executed = true

// Error:
// only owners can create open boards
