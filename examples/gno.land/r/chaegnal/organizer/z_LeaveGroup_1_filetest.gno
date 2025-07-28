package main

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/chaegnal/organizer"
)

var (
	owner = testutils.TestAddress("Chaegnal")
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/chaegnal/organizer"))
}

// Verify error case
// Owner of the group tries to leave the group
func main() {
	organizer.CreateFriendGroup(cross, "Chaegnal's Leave")

	organizer.LeaveGroup(cross, "1")
}

// Error:
// owner cannot leave the group
