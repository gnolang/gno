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
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/chaegnal/organizer/group"))
}

// Test the error case when anyone other than the owner tries to delete the group
func main() {
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	group.DeleteFriendGroup(cross, "1", "Ghhv")
}

// Error:
// only the owner can delete the group
