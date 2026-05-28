// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_change_member_role_06_filetest

package z_change_member_role_06_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	admin address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var bid boards.ID // Operate on board DAO

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test123", false, false)

	boards2.InviteMember(cross(cur), bid, admin, boards2.RoleGuest)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.ChangeMemberRole(cross(cur), bid, "foo", boards2.RoleModerator)
}

// Error:
// invalid member address: foo
