package governance

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/p/nt/commondao"
	"gno.land/r/volos/gov/xvls"
)

func TestVote_NoVotes(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")
	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 1000)
		AddMember(cross, alice)
	})

	var proposal *commondao.Proposal
	crossThrough(std.NewUserRealm(alice), func() {
		proposal = CreateProposal(cross, "No Votes", "Body", time.Second*10, func() {})
	})
	urequire.NotEqual(t, nil, proposal)

	pass, err := proposal.Definition().(VolosProposalDefinition).Tally(proposal.VotingRecord().Readonly(), MemberSet())
	urequire.NoError(t, err)
	uassert.False(t, pass, "proposal should not pass with no votes")
}

func TestVote_AllAbstain(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")
	bob := std.DerivePkgAddr("gno.land/r/volos/gov/bob")
	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 1000)
		xvls.Mint(cross, bob, 1000)
		AddMember(cross, alice)
		AddMember(cross, bob)
	})
	var proposal *commondao.Proposal
	crossThrough(std.NewUserRealm(alice), func() {
		proposal = CreateProposal(cross, "All Abstain", "Body", time.Second*10, func() {})
		Vote(cross, proposal.ID(), "ABSTAIN", "")
	})
	crossThrough(std.NewUserRealm(bob), func() {
		Vote(cross, proposal.ID(), "ABSTAIN", "")
	})
	pass, err := proposal.Definition().(VolosProposalDefinition).Tally(proposal.VotingRecord().Readonly(), MemberSet())
	urequire.NoError(t, err)
	uassert.False(t, pass, "proposal should not pass with all abstain votes")
}

func TestVote_DuplicateAndOverwrite(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")
	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 1000)
		AddMember(cross, alice)
	})
	var proposal *commondao.Proposal
	crossThrough(std.NewUserRealm(alice), func() {
		proposal = CreateProposal(cross, "Duplicate Vote", "Body", time.Second*10, func() {})
		Vote(cross, proposal.ID(), "YES", "First vote")
		Vote(cross, proposal.ID(), "NO", "Changed mind")
	})
	record := proposal.VotingRecord().Readonly()
	vote, found := record.GetVote(alice)
	urequire.True(t, found)
	urequire.Equal(t, "NO", string(vote.Choice))
}

func TestVote_InvalidChoice(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")
	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 1000)
		AddMember(cross, alice)
	})
	var proposal *commondao.Proposal
	crossThrough(std.NewUserRealm(alice), func() {
		proposal = CreateProposal(cross, "Invalid Choice", "Body", time.Second*10, func() {})
		uassert.AbortsWithMessage(t, "invalid vote choice", func() {
			Vote(cross, proposal.ID(), "MAYBE", "not allowed")
		})
	})
}

func TestVote_NonMember(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")
	bob := std.DerivePkgAddr("gno.land/r/volos/gov/bob")
	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 1000)
		AddMember(cross, alice)
	})

	var proposal *commondao.Proposal
	crossThrough(std.NewUserRealm(alice), func() {
		proposal = CreateProposal(cross, "NonMember Vote", "Body", time.Second*10, func() {})
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Burn(cross, bob, xvls.BalanceOf(bob))
		RemoveMember(cross, bob)
	})

	crossThrough(std.NewUserRealm(bob), func() {
		uassert.AbortsWithMessage(t, "account is not a member of the DAO", func() {
			Vote(cross, proposal.ID(), "YES", "not a member")
		})
	})
}

func TestVote_AfterDeadline(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")
	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 1000)
		AddMember(cross, alice)
	})

	var proposal *commondao.Proposal
	crossThrough(std.NewUserRealm(alice), func() {
		proposal = CreateProposal(cross, "After Deadline", "Body", time.Second*1, func() {})
	})

	testing.SkipHeights(2000000)
	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "voting deadline not met", func() {
			Vote(cross, proposal.ID(), "YES", "too late")
		})
	})
}

func TestVote_OnFinishedProposal(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")
	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 1000)
		AddMember(cross, alice)
	})
	var proposal *commondao.Proposal
	crossThrough(std.NewUserRealm(alice), func() {
		proposal = CreateProposal(cross, "Finished Proposal", "Body", time.Second*1, func() {})
		Vote(cross, proposal.ID(), "YES", "For!")
	})
	testing.SkipHeights(2000)
	crossThrough(std.NewUserRealm(alice), func() {
		volosGovernance.Execute(proposal.ID())
	})
	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "proposal not found", func() {
			Vote(cross, proposal.ID(), "YES", "too late")
		})
	})
}
