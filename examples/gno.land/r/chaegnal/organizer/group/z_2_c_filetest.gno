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
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
}

// Test the error case whe the chosen futur owner is not an admin
func main() {
	testing.SetOriginCaller(owner)
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller(user)
	organizer.AskJoinFriendGroup(cross, 1)

	group.AcceptAwaitingMember(cross, "1", owner, std.Address(user))

	testing.SetOriginCaller(owner)
	group.SwitchOwner(cross, "1", owner, user)
}

// Error:
// the new owner must be an admin
