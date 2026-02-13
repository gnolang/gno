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

// Test the correct retrieval of group via it's ID
func main() {
	organizer.CreateFriendGroup(cross, "Test")

	group := organizer.GetGroupFromId("1")

	if group.ID == 1 {
		println("PASS: correct group ID")
	}
	if group.Title == "Test" {
		println("PASS: correct group title")
	}
}

// Output:
// PASS: correct group ID
// PASS: correct group title
