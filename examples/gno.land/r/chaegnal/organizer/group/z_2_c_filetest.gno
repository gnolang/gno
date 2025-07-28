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
	testing.SetRealm(std.NewCodeRealm("gno.land/r/chaegnal/organizer/group"))
}

// Test the error case whe the chosen futur owner is not an admin
func main() {
	testing.SetRealm(std.NewUserRealm(owner))
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetRealm(std.NewUserRealm(user))
	organizer.AskJoinFriendGroup(cross, 1)

	// group.AcceptAwaitingMember(cross, "1", owner.String(), std.Address(user))
	//
	// testing.SetRealm(std.NewUserRealm(owner))
	// group.SwitchOwner(cross, "1", owner.String(), user)
}

// Error:
// the new owner must be an admin
