package organizer

import (
	"testing"
)

// Test getGroupFromId
func Test_getGroupFromId_invalidID(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Invalid ID" {
			t.Errorf("Expected 'Invalid ID' panic")
		}
	}()

	GetGroupFromId("abc")
}

func Test_getGroupFromId_notFound(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Could not find group" {
			t.Errorf("Expected 'Could not find group' panic")
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

	raw, found := ListGroups.Groups.Get("2")
	if !found {
		t.Fatalf("Group was not created")
	}

	group := raw.(FriendGroup)
	if group.ID != 2 {
		t.Errorf("Expected 2 -> got %d", group.ID)
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
		if r != "Group does not exist" {
			t.Errorf("Expected 'Group does not exist' panic")
		}
	}()

	AskJoinFriendGroup(100)
}

func Test_AskJoinFriendGroup_AlreadyPart(t *testing.T) {
	defer func() {
		r := recover()
		if r != "You are already part of this FriendGroup" {
			t.Errorf("Expected 'You are already part of this FriendGroup' panic")
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	CreateFriendGroup("Chaegnal's Already In")

	AskJoinFriendGroup(3)
}

// Test Login / Logout
func Test_Login_Logout(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")

	Login()
	if AddressLoggedIn != "Chaegnal" {
		t.Errorf("Expected 'AddressLoggedIn' to be 'Chaegnal'")
	}

	Logout()
	if AddressLoggedIn != "" {
		t.Errorf("Expected 'AddressLoggedIn' to be empty")
	}
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
	group.DAO.WhitelistingAwaitingApproval("Chaegnal", "Ghhv")

	Login()
	LeaveGroup("4")
	Logout()

	if group.DAO.Whitelist.Has("Ghhv") {
		t.Errorf("'Ghhv' still in whitelist")
	}
}

func Test_LeaveGroup_Failure(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Owner cannot leave the group." {
			t.Errorf("Expected 'Owner cannot leave the group.' panic")
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	CreateFriendGroup("Chaegnal's Leave")

	Login()
	LeaveGroup("4")
	Logout()
}
