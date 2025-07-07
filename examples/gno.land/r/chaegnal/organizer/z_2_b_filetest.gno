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

// Verify an error case
// When the requested group to join does not exist
func main() {
	organizer.AskJoinFriendGroup(cross, 1)
}

// Error:
// group does not exist
