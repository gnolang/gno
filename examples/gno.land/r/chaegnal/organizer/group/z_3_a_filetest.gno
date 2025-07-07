package main

import (
	"std"
	"testing"

	"gno.land/p/demo/seqid"
	"gno.land/r/chaegnal/organizer"
	"gno.land/r/chaegnal/organizer/group"
)

const (
	owner = "Chaegnal"
	user  = "Ghhv"
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/chaegnal/organizer"))
}

// Verify the correct execution of the demotion of an admin
func main() {
	testing.SetOriginCaller(owner)
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller(user)
	organizer.AskJoinFriendGroup(cross, 1)
	group.AcceptAwaitingMember(cross, "1", owner, std.Address(user))

	testing.SetOriginCaller(owner)
	group.MakeAdmin(cross, "1", owner, user)
	group.RemoveAdminRank(cross, "1", owner, user)

	raw, _ := organizer.ListGroups.Groups.Get(seqid.ID(1).String())
	retrieved := raw.(organizer.FriendGroup)
	rank, _ := retrieved.Rank.Get(user)
	if rank.(int) != organizer.AdminRank {
		println("PASS: admin successfully demoted")
	}
}

// Output:
// PASS: admin successfully demoted
