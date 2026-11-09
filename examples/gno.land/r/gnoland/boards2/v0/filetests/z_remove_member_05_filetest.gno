// PKGPATH: gno.land/r/gnoland/boards2/v0/filetests/z_remove_member_05_filetest

package z_remove_member_05_filetest

import (
	"chain/runtime"
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
	// The ephemeral realm MsgRun deploys for the member. Its pkgpath makes it
	// code, and its address is the member's own, which is what carries the
	// self-removal shortcut.
	run := runtime.NewRealm(member, "gno.land/e/"+member.String()+"/run")
	println("caller is code:", run.IsCode())
	println("caller address is the member:", run.Address() == member)

	testing.SetOriginCaller(member)
	testing.SetRealm(run)

	boards2.RemoveMember(cross(cur), bid, member)

	println("still a member:", boards2.IsMember(bid, member))
}

// Output:
// caller is code: true
// caller address is the member: true
// still a member: false
