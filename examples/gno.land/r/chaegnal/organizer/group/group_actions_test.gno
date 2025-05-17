package group

import (
	"testing"

	"gno.land/r/chaegnal/organizer"
)

// Test DeleteFriendGroup
func Test_DeleteFriendGroup_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	DeleteFriendGroup("1", "Chaegnal")
}

func Test_DeleteFriendGroup_Failure(t *testing.T) {
	defer func() {
		r := recover()
		if r != "only the owner can delete the group" {
			t.Errorf("Expected 'only the owner can delete the group' panic, got '%v' panic", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	DeleteFriendGroup("2", "Ghhv")
}

// Test SwitchOwner
func Test_SwitchOwner_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	organizer.AskJoinFriendGroup(3)

	group := organizer.GetGroupFromId("3")
	group.DAO.WhitelistUser("Chaegnal", "Ghhv")

	testing.SetOriginCaller("Chaegnal")
	MakeAdmin("3", "Chaegnal", "Ghhv")
	SwitchOwner("3", "Chaegnal", "Ghhv")
}

func Test_SwitchOwner_NotOwner(t *testing.T) {
	defer func() {
		r := recover()
		if r != "only the owner give ownership away" {
			t.Errorf("Expected 'only the owner give ownership away' panic, got '%v' panic", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	SwitchOwner("4", "Ghhv", "Ghhv")
}

func Test_SwitchOwner_NotAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "the new owner must be an admin" {
			t.Errorf("Expected 'the new owner must be an admin' panic, got '%v' panic", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	organizer.AskJoinFriendGroup(5)

	group := organizer.GetGroupFromId("5")
	group.DAO.WhitelistUser("Chaegnal", "Ghhv")

	testing.SetOriginCaller("Chaegnal")
	SwitchOwner("5", "Chaegnal", "Ghhv")
}
