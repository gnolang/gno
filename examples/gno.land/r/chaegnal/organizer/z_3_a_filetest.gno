package main

import (
	"std"
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

// Verify the correct execution of the process to leave a friend group
func main() {
	organizer.CreateFriendGroup(cross, "Chaegnal's Leave Ghhv")

	testing.SetRealm(std.NewUserRealm(user))
	organizer.AskJoinFriendGroup(cross, 1)

	group := organizer.GetGroupFromId("1")
	adr := std.Address(user)

	group.WaitList.Remove(adr.String())
	group.DAO.Members().Add(adr)
	group.Rank.Set(adr.String(), organizer.MemberRank)

	organizer.LeaveGroup(cross, "1")

	if !group.DAO.Members().Has(user) {
		println("PASS: 'Ghhv' was removed from group")
	}
}

// Output:
// PASS: 'Ghhv' was removed from group
