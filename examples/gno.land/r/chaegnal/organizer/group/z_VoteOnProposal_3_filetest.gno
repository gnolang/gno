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

// Test when a user tries to vote with anything else than yay or nay
func main() {
	testing.SetRealm(std.NewUserRealm(owner))
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	group.CreateProposals(cross, "1", "test", "test", "", "", owner.String())
	group.VoteOnProposal(cross, "1", "1", "2", owner.String())

}

// Error:
// vote should be yay (1) or nay (0)
