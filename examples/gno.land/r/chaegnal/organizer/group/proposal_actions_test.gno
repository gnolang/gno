package group

import (
	"testing"

	"gno.land/r/chaegnal/organizer"
)

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

	CreateProposals("7", "test", "test", "", "", "Chaegnal")
	VoteOnProposal("7", "1", "1", "Chaegnal")
}

func Test_VoteOnProposal_InvalidVoteId(t *testing.T) {
	defer func() {
		r := recover()
		if r != "invalid vote id" {
			t.Errorf("Expected 'invalid vote id' panic, got '%v' panic", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	VoteOnProposal("7", "abc", "1", "Chaegnal")
}

func Test_VoteOnProposal_InvalidVoteChoice(t *testing.T) {
	defer func() {
		r := recover()
		if r != "invalid vote choice" {
			t.Errorf("Expected 'invalid vote choice' panic, got '%v' panic", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	CreateProposals("7", "test", "test", "", "", "Chaegnal")
	VoteOnProposal("7", "1", "abc", "Chaegnal")
}

func Test_VoteOnProposal_NorYayNorNay(t *testing.T) {
	defer func() {
		r := recover()
		if r != "vote should be yay (1) or nay (0)" {
			t.Errorf("Expected 'vote should be yay (1) or nay (0)' panic, got '%v' panic", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	CreateProposals("7", "test", "test", "", "", "Chaegnal")
	VoteOnProposal("7", "1", "2", "Chaegnal")
}

// Test AdminCloseProposal
func Test_AdminCloseProposal_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	CreateProposals("8", "test", "test", "", "", "Chaegnal")
	AdminCloseProposal("8", "1", "Chaegnal")
}

func Test_AdminCloseProposal_Failure(t *testing.T) {
	defer func() {
		r := recover()
		if r != "invalid vote id" {
			t.Errorf("Expected 'invalid vote id' panic, got '%v' panic", r)
		}
	}()
	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup("Chaegnal's group")

	CreateProposals("9", "test", "test", "", "", "Chaegnal")
	AdminCloseProposal("9", "a", "Chaegnal")
}
