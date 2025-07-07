package main

import (
	"std"
	"testing"

	"gno.land/r/chaegnal/organizer"
)

func init() {
	testing.SetOriginCaller("Chaegnal")
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
