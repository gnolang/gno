package dao

import (
	"std"
	"testing"
)

// Test DAO creation
func Test_NewDAO(t *testing.T) {
	creator := std.Address("Chaegnal")
	dao := NewDAO(creator)

	if !dao.IsAdmin(creator) {
		t.Errorf("Creator was not correctly set as admin.")
	}
}

// Test PendingApproval
func Test_PendingApproval_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
}

func Test_PendingApproval_Failure_AlreadyAwaiting(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Already awaiting approval." {
			t.Errorf("Expected 'Already awaiting approval.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.PendingApproval("user")
}

func Test_PendingApproval_Failure_AlreadyIn(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Already whitelisted." {
			t.Errorf("Expected 'Already whitelisted.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("admin")
}

// Test WhitelistingAwaitingApproval
func Test_WhitelistingAwaitingApproval_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("admin", "user")

	member, ok := dao.IsMember("user")
	if !ok || member.Rank != memberRank {
		t.Errorf("Expected user to be whitelisted with member rank.")
	}
}

func Test_WhitelistingAwaitingApproval_Failure_NotAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Only admins can whitelist members." {
			t.Errorf("Expected 'Only admins can whitelist members.' panic")
		}
	}()

	dao := NewDAO("admin")
	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("user", "user")
}

func Test_WhitelistingAwaitingApproval_Failure_NotAwaiting(t *testing.T) {
	defer func() {
		r := recover()
		if r != "This address is not awaiting approval." {
			t.Errorf("Expected 'This address is not awaiting approval.' panic")
		}
	}()

	dao := NewDAO("admin")
	dao.WhitelistingAwaitingApproval("admin", "user")
}

func Test_WhitelistingAwaitingApproval_Failure_AlreadyIn(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Already whitelisted." {
			t.Errorf("Expected 'Already whitelisted.' panic")
		}
	}()

	dao := NewDAO("admin")
	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("admin", "user")
	dao.PendingApproval("user")
}

// Test RefuseAwaitingMember
func Test_RefuseAwaitingMember_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.RefuseAwaitingMember("admin", "user")
}

func Test_RefuseAwaitingMember_Failure(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Only admins can refuse members." {
			t.Errorf("Expected 'Only admins can refuse members.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.RefuseAwaitingMember("user", "user")
}

// Test SetAdmin
func Test_SetAdmin_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("admin", "user")

	dao.SetAdmin("admin", "user")
}

func Test_SetAdmin_Failure_NotAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Only admins can promote members." {
			t.Errorf("Expected 'Only admins can promote members.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("admin", "user")

	dao.SetAdmin("user", "user")
}

// Test defaultAdmin
func Test_SetAdmin_Failure_NotMember(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Not a member." {
			t.Errorf("Expected 'Not a member.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.defaultAdmin("user")
}

func Test_SetAdmin_Failure_AlreadyAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Member is already an admin." {
			t.Errorf("Expected 'Member is already an admin.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("admin", "user")

	dao.defaultAdmin("user")
	dao.defaultAdmin("user")
}

// Test RemoveAdminRank
func Test_RemoveAdminRank_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("admin", "user")
	dao.SetAdmin("admin", "user")

	dao.RemoveAdminRank("admin", "user")
}

func Test_RemoveAdminRank_Failure_CallerNotAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Only admins can demote members." {
			t.Errorf("Expected 'Only admins can demote members.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("admin", "user")

	dao.RemoveAdminRank("user", "user")
}

func Test_RemoveAdminRank_Failure_NotAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Member is not an admin." {
			t.Errorf("Expected 'Member is not an admin.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("admin", "user")

	dao.RemoveAdminRank("admin", "user")
}

func Test_RemoveAdminRank_Failure_AutoDemotion(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Cannot demote yourself." {
			t.Errorf("Expected 'Cannot demote yourself.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.RemoveAdminRank("admin", "admin")
}

// Test RemoveMember
func Test_RemoveMember_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("admin", "user")

	dao.RemoveMember("admin", "user")
}

func Test_RemoveMember_Failure_NotAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Must be an admin to remove a member." {
			t.Errorf("Expected 'Must be an admin to remove a member.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("admin", "user")

	dao.RemoveMember("user", "user")
}

func Test_RemoveMember_Failure_NotMember(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Not a member." {
			t.Errorf("Expected 'Not a member.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.RemoveMember("admin", "user")
}

func Test_RemoveMember_Failure_AutoRemoval(t *testing.T) {
	defer func() {
		r := recover()
		if r != "You cannot remove yourself." {
			t.Errorf("Expected 'You cannot remove yourself.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.RemoveMember("admin", "admin")
}

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
	dao.WhitelistingAwaitingApproval("admin", "user")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	dao.VoteProposal(1, 1, "user")
}

func Test_VoteProposal_Failure_NotMember(t *testing.T) {
	defer func() {
		r := recover()
		if r != "You are not whitelisted." {
			t.Errorf("Expected 'You are not whitelisted.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	dao.VoteProposal(1, 1, "user")
}

func Test_VoteProposal_Failure_InvalidProposal(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Proposal does not exist." {
			t.Errorf("Expected 'Proposal does not exist.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("admin", "user")

	dao.VoteProposal(1, 1, "user")
}

func Test_VoteProposal_Failure_AlreadyVoted(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Already voted." {
			t.Errorf("Expected 'Already voted.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("admin", "user")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	dao.VoteProposal(1, 1, "user")
	dao.VoteProposal(1, 1, "user")
}

func Test_VoteProposal_Failure_InvalidVote(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Invalid vote. Must be 0 (Nay) or 1 (Yay)." {
			t.Errorf("Expected 'Invalid vote. Must be 0 (Nay) or 1 (Yay).' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("admin", "user")

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
		if r != "Proposal does not exist." {
			t.Errorf("Expected 'Proposal does not exist.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.closeProposal(100, "admin")
}

func Test_closeProposal_AlreadyClosed(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Proposal already closed." {
			t.Errorf("Expected 'Proposal already closed.' panic")
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
		if r != "You cannot close a proposal." {
			t.Errorf("Expected 'You cannot close a proposal.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.CreateProposals("Test Title", "Test Description", "testCallback", []string{"arg1", "arg2"}, "admin")

	dao.AdminCloseProposal(1, "user")
}

// Test LeaveDao
func Test_LeaveDao_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistingAwaitingApproval("admin", "user")

	dao.LeaveDao("user")
}

func Test_LeaveDao_Failure_NotMember(t *testing.T) {
	defer func() {
		r := recover()
		if r != "Not a member." {
			t.Errorf("Expected 'Not a member.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.LeaveDao("user")
}

func Test_LeaveDao_Failure_LastAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "You are the last admin. Assign another admin before leaving." {
			t.Errorf("Expected 'You are the last admin. Assign another admin before leaving.' panic")
		}
	}()

	dao := NewDAO("admin")

	dao.LeaveDao("admin")
}
