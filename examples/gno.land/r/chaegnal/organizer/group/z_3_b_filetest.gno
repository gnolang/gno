package main

import (
	"std"
	"testing"

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

// Test the error case when an admin tries to strip the owner of their admin power
func main() {
	testing.SetRealm(std.NewUserRealm(owner))
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetRealm(std.NewUserRealm(user))
	organizer.AskJoinFriendGroup(cross, 1)
	group.AcceptAwaitingMember(cross, "1", owner.String(), std.Address(user))

	testing.SetRealm(std.NewUserRealm(owner))
	group.MakeAdmin(cross, "1", owner.String(), user)

	testing.SetRealm(std.NewUserRealm(user))
	group.RemoveAdminRank(cross, "1", user.String(), owner)
}

// Error:
// you cannot undo admin rights of the owner
