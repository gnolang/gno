package group

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/r/chaegnal/organizer"
)

// RemoveAdminRank
func Test_RemoveAdminRank_Success(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	organizer.AskJoinFriendGroup(cross, 7)

	AcceptAwaitingMember(cross, "7", "Chaegnal", std.Address("Ghhv"))

	testing.SetOriginCaller("Chaegnal")
	MakeAdmin(cross, "7", "Chaegnal", "Ghhv")

	uassert.NotPanics(t, func() { RemoveAdminRank(cross, "7", "Chaegnal", "Ghhv") })
}

func Test_RemoveAdminRank_Failure(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	organizer.AskJoinFriendGroup(cross, 8)

	AcceptAwaitingMember(cross, "8", "Chaegnal", std.Address("Ghhv"))

	testing.SetOriginCaller("Chaegnal")
	MakeAdmin(cross, "8", "Chaegnal", "Ghhv")

	testing.SetOriginCaller("Ghhv")

	err := revive(func() {
		RemoveAdminRank(cross, "8", "Ghhv", "Chaegnal")
	})

	errorMsg := "you cannot undo admin rights of the owner"

	if err == nil || err.(string) != errorMsg {
		t.Fatalf("unexpected panic message: got %q, want %q", err.(string), errorMsg)
	}
}
