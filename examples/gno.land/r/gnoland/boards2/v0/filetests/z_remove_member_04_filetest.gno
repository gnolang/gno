// PKGPATH: gno.land/r/gnoland/boards2/v0/filetests/z_remove_member_04_filetest

package z_remove_member_04_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards/v0"

	boards2 "gno.land/r/gnoland/boards2/v0"
)

const (
	owner  address = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"
	member address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var bid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test123", false, false)
	boards2.InviteMember(cross(cur), bid, member, boards2.RoleGuest)
}

func main(cur realm) {
	// The member's address is the transaction signer (OriginCaller), but the
	// immediate caller is a different code realm.
	testing.SetOriginCaller(member)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/v0/filetests/z_remove_member_04_filetest"))

	// Self-removal keys off the immediate caller, not the signer. This realm
	// is not the member, so RemoveMember falls through to the permission
	// check, which this realm lacks, and the call panics.
	boards2.RemoveMember(cross(cur), bid, member)
}

// Error:
// unauthorized, user g1cg89vd3vkavgd72r95gsdsjwc2wy6qy4vggas9 doesn't have the required permission
