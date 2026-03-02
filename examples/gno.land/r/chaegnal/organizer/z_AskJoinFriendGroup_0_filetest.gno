package main

import (
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/chaegnal/organizer"
)

var (
	owner = testutils.TestAddress("Chaegnal")
	user  = testutils.TestAddress("Ghhv")
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/chaegnal/organizer"))
}

// Verify the correct execution of asking to join a group
func main() {
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetRealm(std.NewUserRealm(user))
	organizer.AskJoinFriendGroup(cross, 1)

	group := organizer.GetGroupFromId(strconv.Itoa(1))
	if group.WaitList.Has(user.String()) {
		println("PASS: 'Ghhv' is in the waitlist")
	}
}

// Output:
// PASS: 'Ghhv' is in the waitlist
