package dao

import (
	"testing"

	"gno.land/p/demo/uassert"
)

// Test CreateProposals
func Test_CreateProposals(t *testing.T) {
	dao := NewDAO("admin")

	uassert.NotPanics(t, func() {
		dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")
	})

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
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	uassert.NotPanics(t, func() { dao.VoteProposal(1, 1, "user") })
}

func Test_VoteProposal_Failure_NotMember(t *testing.T) {
	dao := NewDAO("admin")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	uassert.PanicsWithMessage(t, "you are not whitelisted", func() { dao.VoteProposal(1, 1, "user") })
}

func Test_VoteProposal_Failure_InvalidProposal(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	uassert.PanicsWithMessage(t, "proposal does not exist", func() { dao.VoteProposal(1, 1, "user") })
}

func Test_VoteProposal_Failure_AlreadyVoted(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	dao.VoteProposal(1, 1, "user")
	uassert.PanicsWithMessage(t, "already voted", func() { dao.VoteProposal(1, 1, "user") })
}

func Test_VoteProposal_Failure_InvalidVote(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	uassert.PanicsWithMessage(t, "invalid vote: must be 0 (nay) or 1 (yay)", func() { dao.VoteProposal(1, 2, "user") })
}

func Test_closeProposal_Success(t *testing.T) {
	dao := NewDAO("admin")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	uassert.NotPanics(t, func() { dao.closeProposal(1, "admin") })
}

func Test_closeProposal_NoExist(t *testing.T) {
	dao := NewDAO("admin")

	uassert.PanicsWithMessage(t, "proposal does not exist", func() { dao.closeProposal(100, "admin") })
}

func Test_closeProposal_AlreadyClosed(t *testing.T) {
	dao := NewDAO("admin")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	dao.closeProposal(1, "admin")
	uassert.PanicsWithMessage(t, "proposal already closed", func() { dao.closeProposal(1, "admin") })
}

// Test AdminCloseProposal
func Test_AdminCloseProposal_Success(t *testing.T) {
	dao := NewDAO("admin")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	uassert.NotPanics(t, func() { dao.AdminCloseProposal(1, "admin") })
}

func Test_AdminCloseProposal_Failure(t *testing.T) {
	dao := NewDAO("admin")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	uassert.PanicsWithMessage(t, "you cannot close a proposal", func() { dao.AdminCloseProposal(1, "user") })
}
