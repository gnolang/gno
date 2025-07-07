package main

import (
	"std"
	"strconv"
	"testing"

	"gno.land/r/chaegnal/organizer"
)

const (
	owner = "Chaegnal"
	user  = "Ghhv"
)

func init() {
	testing.SetOriginCaller(owner)
	testing.SetRealm(std.NewCodeRealm("gno.land/r/chaegnal/organizer"))
}

// Verify the correct execution of asking to join a group
func main() {
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller(user)
	organizer.AskJoinFriendGroup(cross, 1)

	group := organizer.GetGroupFromId(strconv.Itoa(1))
	if group.WaitList.Has(user) {
		println("PASS: 'Ghhv' is in the waitlist")
	}
}

// Output:
// PASS: 'Ghhv' is in the waitlist
