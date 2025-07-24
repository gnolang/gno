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

// Verify the correct execution of voting on a proposal
func main() {
	testing.SetRealm(std.NewUserRealm(owner))
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	group.CreateProposals(cross, "1", "test", "test", "", "", owner.String())
	group.VoteOnProposal(cross, "1", "1", "1", owner.String())
}

// Error:
