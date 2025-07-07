package main

import (
	"std"
	"testing"

	"gno.land/p/demo/seqid"
	"gno.land/r/chaegnal/organizer"
)

func init() {
	testing.SetOriginCaller("Chaegnal")
	testing.SetRealm(std.NewCodeRealm("gno.land/r/chaegnal/organizer"))
}

// Verifies the correct creation of a friend group
func main() {
	organizer.CreateFriendGroup(cross, "Test")

	raw, found := organizer.ListGroups.Groups.Get(seqid.ID(1).String())
	if !found {
		println("FAIL: group was not created")
		return
	}

	group := raw.(organizer.FriendGroup)

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
