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

// Verify the error case when someone else than the owner tries to change ownership
func main() {
	testing.SetOriginCaller(owner)
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller(user)
	group.SwitchOwner(cross, "1", user, user)
}

// Error:
// only the owner give ownership away
