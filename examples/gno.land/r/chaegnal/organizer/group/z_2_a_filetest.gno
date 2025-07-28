package main

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/chaegnal/organizer"
)

var (
	owner = testutils.TestAddress("Chaegnal")
	user  = testutils.TestAddress("Ghhv")
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/chaegnal/organizer/group"))
}

// Verify the correct execution of ownership handover
func main() {
	testing.SetRealm(std.NewUserRealm(owner))
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetRealm(std.NewUserRealm(user))
	organizer.AskJoinFriendGroup(cross, 1)

	// group.AcceptAwaitingMember(cross, "1", owner.String(), std.Address(user))
	//
	// testing.SetRealm(std.NewUserRealm(owner))
	// group.MakeAdmin(cross, "1", owner.String(), user)
	//
	// group.SwitchOwner(cross, "1", owner.String(), user)
	//
	// retrieved := organizer.GetGroupFromId(strconv.Itoa(1))
	// if retrieved.Owner.Owner() == user {
	// println("PASS: owner successfully changed")
	// }
}

// Output:
// PASS: owner successfully changed
