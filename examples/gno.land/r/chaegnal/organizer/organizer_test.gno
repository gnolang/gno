package organizer

import (
	"testing"
)

// Test CreateFriendGroup
func Test_CreateFriendGroup(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")

	CreateFriendGroup("Test")

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
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	CreateFriendGroup("Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	AskJoinFriendGroup(2)

	group := GetGroupFromId("2")
	if !group.DAO.IsPending("Ghhv") {
		t.Errorf("'Ghhv' not in waitlist")
	}
}

func Test_AskJoinFriendGroup_GroupFail(t *testing.T) {
	defer func() {
		r := recover()
		if r != "group does not exist" {
			t.Errorf("Expected 'group does not exist' panic, got '%v' panic", r)
		}
	}()

	AskJoinFriendGroup(100)
}

func Test_AskJoinFriendGroup_AlreadyPart(t *testing.T) {
	defer func() {
		r := recover()
		if r != "you are already part of this friend group" {
			t.Errorf("Expected 'you are already part of this friend group' panic, got '%v' panic", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	CreateFriendGroup("Chaegnal's Already In")

	AskJoinFriendGroup(3)
}

// Test LeaveGroup
func Test_LeaveGroup_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	CreateFriendGroup("Chaegnal's Leave Ghhv")

	testing.SetOriginCaller("Ghhv")
	AskJoinFriendGroup(4)

	group := GetGroupFromId("4")
	group.DAO.WhitelistUser("Chaegnal", "Ghhv")

	LeaveGroup("4")

	if group.DAO.Whitelist.Has("Ghhv") {
		t.Errorf("'Ghhv' still in whitelist")
	}
}

func Test_LeaveGroup_Failure(t *testing.T) {
	defer func() {
		r := recover()
		if r != "owner cannot leave the group" {
			t.Errorf("Expected 'owner cannot leave the group' panic, got '%v' panic", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	CreateFriendGroup("Chaegnal's Leave")

	LeaveGroup("4")
}

// Helpers

// Test getGroupFromId
func Test_getGroupFromId_invalidID(t *testing.T) {
	defer func() {
		r := recover()
		if r != "invalid id" {
			t.Errorf("Expected 'invalid id' panic, got '%v' panic", r)
		}
	}()

	GetGroupFromId("abc")
}

func Test_getGroupFromId_notFound(t *testing.T) {
	defer func() {
		r := recover()
		if r != "could not find group" {
			t.Errorf("Expected 'could not find group' panic, got '%v' panic", r)
		}
	}()

	GetGroupFromId("100")
}

func TestGetGroupFromId_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	CreateFriendGroup("Test")

	group := GetGroupFromId("1")

	if group.Title != "Test" {
		t.Errorf("Expected 'Test' -> got %q", group.Title)
	}
	if group.ID != 1 {
		t.Errorf("Expected 1 -> got %d", group.ID)
	}
}
