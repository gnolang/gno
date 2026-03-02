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

func main() {
	organizer.CreateFriendGroup(cross, "Chaegnal's Already In")

	organizer.AskJoinFriendGroup(cross, 1)
}

// Error:
// you are already part of this friend group
