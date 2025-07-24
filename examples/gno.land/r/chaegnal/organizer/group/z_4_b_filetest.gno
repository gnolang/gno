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
	testing.SetRealm(std.NewCodeRealm("gno.land/r/chaegnal/organizer/group"))
}

// Test invalid vote id
func main() {
	testing.SetRealm(std.NewUserRealm(owner))
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	group.VoteOnProposal(cross, "1", "abc", "1", owner.String())
}

// Error:
// invalid vote id
