package main

import (
	"std"
	"strconv"
	"testing"

	"gno.land/r/chaegnal/organizer"
	"gno.land/r/chaegnal/organizer/group"
)

func init() {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/chaegnal/organizer/group"))
}

// Test the correct deletion of a friend group by the owner
func main() {
	testing.SetRealm(std.NewUserRealm("Chaegnal"))
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	group.DeleteFriendGroup(cross, "1", "Chaegnal")

	organizer.GetGroupFromId(strconv.Itoa(1))
}

// Error:
// could not find group
