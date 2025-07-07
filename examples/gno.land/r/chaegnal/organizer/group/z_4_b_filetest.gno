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

// Test invalid vote id
func main() {
	testing.SetOriginCaller(owner)
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	group.VoteOnProposal(cross, "1", "abc", "1", owner)
}

// Error:
// invalid vote id
