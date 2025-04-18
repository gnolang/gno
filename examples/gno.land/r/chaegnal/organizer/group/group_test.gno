package group

import (
	"testing"

	"gno.land/r/chaegnal/organizer"
)

// Test getGroupFromUrl
func Test_getGroupFromUrl_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	getGroupFromUrl("http://test.com?groupId=1")
}

func Test_getGroupFromUrl_Failure(t *testing.T) {
	defer func() {
		r := recover()
		if r != "No groupId provided" {
			t.Errorf("Expected 'No groupId provided' panic")
		}
	}()

	getGroupFromUrl("http://test.com?groupId=")
}

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

	organizer.Login()
	DeleteFriendGroup("2")
	organizer.Logout()
}

func Test_DeleteFriendGroup_Failure(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Only the owner can delete the group." {
			t.Errorf("Expected 'Only the owner can delete the group.' panic")
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	DeleteFriendGroup("3")
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
	organizer.AskJoinFriendGroup(4)

	group := organizer.GetGroupFromId("4")
	group.DAO.WhitelistingAwaitingApproval("Chaegnal", "Ghhv")

	testing.SetOriginCaller("Chaegnal")
	organizer.Login()
	MakeAdmin("4", "Ghhv")

	SwitchOwner("4", "Ghhv")
	organizer.Logout()
}

func Test_SwitchOwner_NotOwner(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Only the owner give ownership away." {
			t.Errorf("Expected 'Only the owner give ownership away.' panic")
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	organizer.Login()

	SwitchOwner("5", "Ghhv")
	organizer.Logout()
}

func Test_SwitchOwner_NotAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "The new owner must be an admin." {
			t.Errorf("Expected 'The new owner must be an admin.' panic")
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	organizer.AskJoinFriendGroup(6)

	group := organizer.GetGroupFromId("6")
	group.DAO.WhitelistingAwaitingApproval("Chaegnal", "Ghhv")

	testing.SetOriginCaller("Chaegnal")
	organizer.Login()
	SwitchOwner("6", "Ghhv")
	organizer.Logout()
}

// Test VoteOnProposal
func Test_VoteOnProposal_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	organizer.Login()
	CreateProposals("7", "test", "test", "", "")
	VoteOnProposal("7", "1", "1")
	organizer.Logout()
}

func Test_VoteOnProposal_InvalidVoteId(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Invalid vote ID." {
			t.Errorf("Expected 'Invalid vote ID.' panic")
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	organizer.Login()
	VoteOnProposal("7", "abc", "1")
	organizer.Logout()
}

func Test_VoteOnProposal_InvalidVoteChoice(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Invalid vote choice." {
			t.Errorf("Expected 'Invalid vote choice.' panic")
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	organizer.Login()
	CreateProposals("7", "test", "test", "", "")
	VoteOnProposal("7", "1", "abc")
	organizer.Logout()
}

func Test_VoteOnProposal_NorYayNorNay(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Vote should be Yay (1) or Nay (0)." {
			t.Errorf("Expected 'Vote should be Yay (1) or Nay (0).' panic")
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	organizer.Login()
	CreateProposals("7", "test", "test", "", "")
	VoteOnProposal("7", "1", "2")
	organizer.Logout()
}
