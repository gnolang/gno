// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_invite_member_00_filetest

package z_invite_member_00_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	bid           = boards.ID(0)                               // Operate on realm DAO instead of individual boards
	role          = boards2.RoleOwner
)

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.InviteMember(cross(cur), bid, user, role)

	// Check that user is invited
	println(boards2.HasMemberRole(bid, user, role))
}

// Output:
// true
