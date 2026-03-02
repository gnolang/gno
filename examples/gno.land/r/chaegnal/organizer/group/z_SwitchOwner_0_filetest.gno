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

// Verify the error case when someone else than the owner tries to change ownership
func main() {
	testing.SetRealm(std.NewUserRealm(owner))
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetRealm(std.NewUserRealm(user))
	group.SwitchOwner(cross, "1", user.String(), user)
}

// Error:
// only the owner give ownership away
