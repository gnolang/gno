package main

import (
	"std"
	"testing"

	"gno.land/r/chaegnal/organizer"
	"gno.land/r/chaegnal/organizer/group"
)

const (
	owner = "g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g"
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/chaegnal/organizer/group"))
}

// Test when a user tries to vote with anything else than yay or nay
func main() {
	testing.SetOriginCaller(owner)
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	group.CreateProposals(cross, "1", "test", "test", "", "", owner)
	group.VoteOnProposal(cross, "1", "1", "2", owner)

}

// Error:
// vote should be yay (1) or nay (0)
