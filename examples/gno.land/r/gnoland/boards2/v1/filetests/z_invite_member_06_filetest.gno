// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_invite_member_06_filetest

package z_invite_member_06_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.InviteMember(cross(cur), 0, "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", boards2.RoleGuest)
}

// Error:
// unauthorized, user g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj doesn't have the required permission
