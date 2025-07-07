package main

import (
	"std"
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
	testing.SetRealm(std.NewCodeRealm("gno.land/r/chaegnal/organizer"))
}

// Test the error case when an admin tries to strip the owner of their admin power
func main() {
	testing.SetOriginCaller(owner)
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller(user)
	organizer.AskJoinFriendGroup(cross, 1)
	group.AcceptAwaitingMember(cross, "1", owner, std.Address(user))

	testing.SetOriginCaller(owner)
	group.MakeAdmin(cross, "1", owner, user)

	testing.SetOriginCaller(user)
	group.RemoveAdminRank(cross, "1", user, owner)
}

// Error:
// you cannot undo admin rights of the owner
