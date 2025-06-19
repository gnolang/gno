package organizer

import (
	"testing"

	"gno.land/p/demo/uassert"
)

// Test CreateFriendGroup
func Test_CreateFriendGroup(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")

	uassert.NotPanics(t, func() { CreateFriendGroup(cross, "Test") })

	raw, found := ListGroups.Groups.Get("1")
	if !found {
		t.Fatalf("Group was not created")
	}

	group := raw.(FriendGroup)
	if group.ID != 1 {
		t.Errorf("Expected 1 -> got %d", group.ID)
	}

	if group.Title != "Test" {
		t.Errorf("Expected 'Test', got %q", group.Title)
	}

	if group.Owner.Owner() != "Chaegnal" {
		t.Errorf("Expected 'Chaegnal', got %q", group.Owner.Owner())
	}
}

// Test AskJoinFriendGroup
func Test_AskJoinFriendGroup_Success(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	uassert.NotPanics(t, func() { AskJoinFriendGroup(cross, 2) })

	group := GetGroupFromId("2")
	if !group.DAO.IsPending("Ghhv") {
		t.Errorf("'Ghhv' not in waitlist")
	}
}

func Test_AskJoinFriendGroup_GroupFail(t *testing.T) {
	uassert.PanicsWithMessage(t, "group does not exist", func() { AskJoinFriendGroup(cross, 100) })
}

func Test_AskJoinFriendGroup_AlreadyPart(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	CreateFriendGroup(cross, "Chaegnal's Already In")

	uassert.PanicsWithMessage(t, "you are already part of this friend group", func() { AskJoinFriendGroup(cross, 3) })
}

// Test LeaveGroup
func Test_LeaveGroup_Success(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	CreateFriendGroup(cross, "Chaegnal's Leave Ghhv")

	testing.SetOriginCaller("Ghhv")
	AskJoinFriendGroup(cross, 4)

	group := GetGroupFromId("4")
	group.DAO.WhitelistUser("Chaegnal", "Ghhv")

	uassert.NotPanics(t, func() { LeaveGroup(cross, "4") })

	if group.DAO.Whitelist.Has("Ghhv") {
		t.Errorf("'Ghhv' still in whitelist")
	}
}

func Test_LeaveGroup_Failure(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	CreateFriendGroup(cross, "Chaegnal's Leave")

	uassert.PanicsWithMessage(t, "owner cannot leave the group", func() { LeaveGroup(cross, "4") })
}

// Helpers

// Test getGroupFromId
func Test_getGroupFromId_invalidID(t *testing.T) {
	uassert.PanicsWithMessage(t, "invalid id", func() { GetGroupFromId("abc") })
}

func Test_getGroupFromId_notFound(t *testing.T) {
	uassert.PanicsWithMessage(t, "could not find group", func() { GetGroupFromId("100") })
}

func TestGetGroupFromId_Success(t *testing.T) {
	uassert.NotPanics(t, func() { CreateFriendGroup(cross, "Test") })

	group := GetGroupFromId("1")

	if group.Title != "Test" {
		t.Errorf("Expected 'Test' -> got %q", group.Title)
	}
	if group.ID != 1 {
		t.Errorf("Expected 1 -> got %d", group.ID)
	}
}
