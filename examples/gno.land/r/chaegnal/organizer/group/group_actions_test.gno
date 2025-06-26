package group

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/r/chaegnal/organizer"
)

// Test DeleteFriendGroup
func Test_DeleteFriendGroup_Success(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	uassert.NotPanics(t, func() { DeleteFriendGroup(cross, "1", "Chaegnal") })
}

func Test_DeleteFriendGroup_Failure(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	err := revive(func() {
		DeleteFriendGroup(cross, "2", "Ghhv")
	})

	errorMsg := "only the owner can delete the group"

	if err == nil || err.(string) != errorMsg {
		t.Fatalf("unexpected panic message: got %q, want %q", err.(string), errorMsg)
	}
}

// Test SwitchOwner
func Test_SwitchOwner_Success(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	organizer.AskJoinFriendGroup(cross, 3)

	AcceptAwaitingMember(cross, "3", "Chaegnal", std.Address("Ghhv"))

	testing.SetOriginCaller("Chaegnal")
	MakeAdmin(cross, "3", "Chaegnal", "Ghhv")

	uassert.NotPanics(t, func() { SwitchOwner(cross, "3", "Chaegnal", "Ghhv") })
}

func Test_SwitchOwner_NotOwner(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller("Ghhv")

	err := revive(func() {
		SwitchOwner(cross, "4", "Ghhv", "Ghhv")
	})

	errorMsg := "only the owner give ownership away"

	if err == nil || err.(string) != errorMsg {
		t.Fatalf("unexpected panic message: got %q, want %q", err.(string), errorMsg)
	}
}

func Test_SwitchOwner_NotAdmin(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	organizer.AskJoinFriendGroup(cross, 5)

	AcceptAwaitingMember(cross, "5", "Chaegnal", std.Address("Ghhv"))

	testing.SetOriginCaller("Chaegnal")
	err := revive(func() {
		SwitchOwner(cross, "5", "Chaegnal", "Ghhv")
	})

	errorMsg := "the new owner must be an admin"

	if err == nil || err.(string) != errorMsg {
		t.Fatalf("unexpected panic message: got %q, want %q", err.(string), errorMsg)
	}
}
