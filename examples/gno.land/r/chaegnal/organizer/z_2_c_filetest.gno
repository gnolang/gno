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

func main() {
	organizer.CreateFriendGroup(cross, "Chaegnal's Already In")

	organizer.AskJoinFriendGroup(cross, 1)
}

// Error:
// you are already part of this friend group
