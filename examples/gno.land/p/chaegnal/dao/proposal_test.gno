package dao

import "testing"

// Test CreateProposals
func Test_CreateProposals(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	raw, _ := dao.Proposals.Get("1")
	proposal := raw.(Proposal)

	if proposal.Title != "Test Title" ||
		proposal.Owner != "admin" ||
		proposal.Callback != "testCallback" {
		t.Errorf("Badly formed proposal.")
	}
	if !proposal.Active {
		t.Errorf("Proposal was not activated.")
	}
}

// Test VoteProposal
func Test_VoteProposal_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	dao.VoteProposal(1, 1, "user")
}

func Test_VoteProposal_Failure_NotMember(t *testing.T) {
	defer func() {
		r := recover()
		if r != "you are not whitelisted" {
			t.Errorf("Expected 'you are not whitelisted' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	dao.VoteProposal(1, 1, "user")
}

func Test_VoteProposal_Failure_InvalidProposal(t *testing.T) {
	defer func() {
		r := recover()
		if r != "proposal does not exist" {
			t.Errorf("Expected 'proposal does not exist' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.VoteProposal(1, 1, "user")
}

func Test_VoteProposal_Failure_AlreadyVoted(t *testing.T) {
	defer func() {
		r := recover()
		if r != "already voted" {
			t.Errorf("Expected 'already voted' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	dao.VoteProposal(1, 1, "user")
	dao.VoteProposal(1, 1, "user")
}

func Test_VoteProposal_Failure_InvalidVote(t *testing.T) {
	defer func() {
		r := recover()
		if r != "invalid vote: must be 0 (nay) or 1 (yay)" {
			t.Errorf("Expected 'invalid vote: must be 0 (nay) or 1 (yay)' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	dao.VoteProposal(1, 2, "user")
}

func Test_closeProposal_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	dao.closeProposal(1, "admin")
}

func Test_closeProposal_NoExist(t *testing.T) {
	defer func() {
		r := recover()
		if r != "proposal does not exist" {
			t.Errorf("Expected 'proposal does not exist' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.closeProposal(100, "admin")
}

func Test_closeProposal_AlreadyClosed(t *testing.T) {
	defer func() {
		r := recover()
		if r != "proposal already closed" {
			t.Errorf("Expected 'proposal already closed' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	dao.closeProposal(1, "admin")
	dao.closeProposal(1, "admin")
}

// Test AdminCloseProposal
func Test_AdminCloseProposal_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	dao.AdminCloseProposal(1, "admin")
}

func Test_AdminCloseProposal_Failure(t *testing.T) {
	defer func() {
		r := recover()
		if r != "you cannot close a proposal" {
			t.Errorf("Expected 'you cannot close a proposal' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	dao.AdminCloseProposal(1, "user")
}
