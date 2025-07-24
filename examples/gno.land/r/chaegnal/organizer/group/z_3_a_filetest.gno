package main

import (
	"std"
	"testing"

	"gno.land/p/demo/seqid"
	"gno.land/p/demo/testutils"
	"gno.land/r/chaegnal/organizer"
	"gno.land/r/chaegnal/organizer/group"
)

var (
	owner = testutils.TestAddress("Chaegnal")
	user  = testutils.TestAddress("Ghhv")
)

func init() {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/chaegnal/organizer/group"))
}

// Verify the correct execution of the demotion of an admin
func main() {
	testing.SetRealm(std.NewUserRealm(owner))
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetRealm(std.NewUserRealm(user))
	organizer.AskJoinFriendGroup(cross, 1)
	group.AcceptAwaitingMember(cross, "1", owner.String(), std.Address(user))

	testing.SetRealm(std.NewUserRealm(owner))
	group.MakeAdmin(cross, "1", owner.String(), user)
	group.RemoveAdminRank(cross, "1", owner.String(), user)

	raw, _ := organizer.ListGroups.Groups.Get(seqid.ID(1).String())
	retrieved := raw.(organizer.FriendGroup)
	rank, _ := retrieved.Rank.Get(user.String())
	if rank.(int) != organizer.AdminRank {
		println("PASS: admin successfully demoted")
	}
}

// Output:
// PASS: admin successfully demoted
