package main

import (
	"std"
	"strconv"
	"testing"

	"gno.land/r/chaegnal/organizer"
	"gno.land/r/chaegnal/organizer/group"
)

const (
	owner = "Chaegnal"
	user  = "Ghhv"
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
}

// Verify the correct execution of ownership handover
func main() {
	testing.SetOriginCaller(owner)
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller(user)
	organizer.AskJoinFriendGroup(cross, 1)

	group.AcceptAwaitingMember(cross, "1", owner, std.Address(user))

	testing.SetOriginCaller(owner)
	group.MakeAdmin(cross, "1", owner, user)

	group.SwitchOwner(cross, "1", owner, user)

	retrieved := organizer.GetGroupFromId(strconv.Itoa(1))
	if retrieved.Owner.Owner() == user {
		println("PASS: owner successfully changed")
	}
}

// Output:
// PASS: owner successfully changed
