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

func crossThrough(rlm std.Realm, cr func()) {
	testing.SetRealm(rlm)
	cr()
}

func TestProposal_CreationThreshold(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	gov := "gno.land/r/volos/gov/governance"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 2000)
		AddMember(cross, alice)
	})

	crossThrough(std.NewUserRealm(alice), func() {
		CreateProposal(cross, "Test Proposal", "Body", time.Second*10, func() {})
	})
	urequire.Equal(t, 1, volosGovernance.ActiveProposals().Size())

	crossThrough(std.NewCodeRealm(gov), func() {
		SetProposalThreshold(5000)
	})
	urequire.Equal(t, int64(5000), ProposalThreshold())

	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "insufficient xVLS to propose", func() {
			CreateProposal(cross, "Another Proposal", "Body", time.Second*10, func() {})
		})
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 5000)
	})
	crossThrough(std.NewUserRealm(alice), func() {
		CreateProposal(cross, "Another Proposal", "Body", time.Second*10, func() {})
	})
	urequire.Equal(t, 2, volosGovernance.ActiveProposals().Size())
}

func TestVoting_WeightedAndInvalid(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")
	bob := std.DerivePkgAddr("gno.land/r/volos/gov/bob")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 2000)
		xvls.Mint(cross, bob, 1000)
		AddMember(cross, alice)
		AddMember(cross, bob)
	})

	crossThrough(std.NewUserRealm(alice), func() {
		CreateProposal(cross, "Voting Test", "Body", time.Second*10, func() {})
	})

	crossThrough(std.NewUserRealm(bob), func() {
		uassert.AbortsWithMessage(t, "invalid vote choice", func() {
			Vote(cross, 1, "INVALID", "bad choice")
		})
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Burn(cross, bob, 1000)
		RemoveMember(cross, bob)
	})
	crossThrough(std.NewUserRealm(bob), func() {
		uassert.AbortsWithMessage(t, "account is not a member of the DAO", func() {
			Vote(cross, 1, "YES", "I want this!")
		})
	})

	crossThrough(std.NewUserRealm(alice), func() {
		Vote(cross, 1, "YES", "For!")
	})

	proposal := volosGovernance.GetProposal(1)
	pass, err := proposal.Definition().(VolosProposalDefinition).Tally(proposal.VotingRecord().Readonly(), MemberSet())
	urequire.NoError(t, err)
	urequire.True(t, pass)
}

func TestGovernance_KickMember(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	gov := "gno.land/r/volos/gov/governance"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 2000)
		AddMember(cross, alice)
	})
	urequire.True(t, volosGovernance.Members().Has(alice))

	crossThrough(std.NewCodeRealm(gov), func() {
		KickMember(cross, alice)
	})
	urequire.False(t, volosGovernance.Members().Has(alice))
}

func TestGovernance_KickAndReaddMember(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	gov := "gno.land/r/volos/gov/governance"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Burn(cross, alice, xvls.BalanceOf(alice))
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 2000)
		AddMember(cross, alice)
	})
	urequire.True(t, volosGovernance.Members().Has(alice))

	crossThrough(std.NewCodeRealm(gov), func() {
		KickMember(cross, alice)
	})
	urequire.False(t, volosGovernance.Members().Has(alice))
	urequire.Equal(t, int64(2000), xvls.BalanceOf(alice))

	crossThrough(std.NewCodeRealm(staker), func() {
		AddMember(cross, alice)
	})
	urequire.True(t, volosGovernance.Members().Has(alice))
}

func TestGovernance_QuorumParameters(cur realm, t *testing.T) {
	gov := "gno.land/r/volos/gov/governance"

	urequire.Equal(t, int64(1000), VotingPowerQuorum())

	crossThrough(std.NewCodeRealm(gov), func() {
		SetVotingPowerQuorum(5000)
	})
	urequire.Equal(t, int64(5000), VotingPowerQuorum())

	crossThrough(std.NewCodeRealm(gov), func() {
		SetVotingPowerQuorum(1000)
	})
	urequire.Equal(t, int64(1000), VotingPowerQuorum())
}

func TestProposal_QuorumEnforcement(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	gov := "gno.land/r/volos/gov/governance"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")
	bob := std.DerivePkgAddr("gno.land/r/volos/gov/bob")
	charlie := std.DerivePkgAddr("gno.land/r/volos/gov/charlie")

	crossThrough(std.NewCodeRealm(gov), func() {
		SetProposalThreshold(100)
		SetVotingPowerQuorum(5000)
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Burn(cross, alice, xvls.BalanceOf(alice))
		xvls.Burn(cross, bob, xvls.BalanceOf(bob))
		xvls.Burn(cross, charlie, xvls.BalanceOf(charlie))
		xvls.Mint(cross, alice, 2000)
		xvls.Mint(cross, bob, 2000)
		xvls.Mint(cross, charlie, 2000)
		AddMember(cross, alice)
		AddMember(cross, bob)
		AddMember(cross, charlie)
	})

	var proposalId uint64
	crossThrough(std.NewUserRealm(alice), func() {
		proposalId = CreateProposal(cross, "Quorum Test", "Body", time.Second*10, func() {}).ID()
	})

	crossThrough(std.NewUserRealm(alice), func() {
		Vote(cross, proposalId, "YES", "For!")
	})
	crossThrough(std.NewUserRealm(bob), func() {
		Vote(cross, proposalId, "YES", "For!")
	})

	proposal := volosGovernance.GetProposal(proposalId)
	pass, err := proposal.Definition().(VolosProposalDefinition).Tally(proposal.VotingRecord().Readonly(), MemberSet())
	urequire.NoError(t, err)
	urequire.False(t, pass, "proposal should fail due to voting power quorum not met")

	crossThrough(std.NewUserRealm(charlie), func() {
		Vote(cross, proposalId, "YES", "For!")
	})

	pass, err = proposal.Definition().(VolosProposalDefinition).Tally(proposal.VotingRecord().Readonly(), MemberSet())
	urequire.NoError(t, err)
	urequire.True(t, pass, "proposal should pass with 3 voters and 60000 voting power")

	crossThrough(std.NewCodeRealm(gov), func() {
		SetVotingPowerQuorum(1000)
	})
}

func TestProposal_VotingPowerQuorumEnforcement(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	gov := "gno.land/r/volos/gov/governance"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")
	bob := std.DerivePkgAddr("gno.land/r/volos/gov/bob")

	crossThrough(std.NewCodeRealm(gov), func() {
		SetVotingPowerQuorum(10000)
		SetProposalThreshold(500)
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Burn(cross, alice, xvls.BalanceOf(alice))
		xvls.Burn(cross, bob, xvls.BalanceOf(bob))
		xvls.Mint(cross, alice, 3000)
		xvls.Mint(cross, bob, 3000)
		AddMember(cross, alice)
		AddMember(cross, bob)
	})

	var proposalId uint64
	crossThrough(std.NewUserRealm(alice), func() {
		proposalId = CreateProposal(cross, "Voting Power Quorum Test", "Body", time.Second*10, func() {}).ID()
	})

	crossThrough(std.NewUserRealm(alice), func() {
		Vote(cross, proposalId, "YES", "For!")
	})
	crossThrough(std.NewUserRealm(bob), func() {
		Vote(cross, proposalId, "YES", "For!")
	})

	proposal := volosGovernance.GetProposal(proposalId)
	pass, err := proposal.Definition().(VolosProposalDefinition).Tally(proposal.VotingRecord().Readonly(), MemberSet())
	urequire.NoError(t, err)
	urequire.False(t, pass, "proposal should fail due to voting power quorum not met")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 5000)
		xvls.Mint(cross, bob, 5000)
	})

	crossThrough(std.NewUserRealm(alice), func() {
		Vote(cross, proposalId, "YES", "For!")
	})
	crossThrough(std.NewUserRealm(bob), func() {
		Vote(cross, proposalId, "YES", "For!")
	})

	pass, err = proposal.Definition().(VolosProposalDefinition).Tally(proposal.VotingRecord().Readonly(), MemberSet())
	urequire.NoError(t, err)
	urequire.True(t, pass, "proposal should pass with sufficient voting power")

	crossThrough(std.NewCodeRealm(gov), func() {
		SetVotingPowerQuorum(1000)
	})
}

func TestGovernance_QuorumEnableDisable(cur realm, t *testing.T) {
	gov := "gno.land/r/volos/gov/governance"

	urequire.Equal(t, int64(1000), VotingPowerQuorum())

	crossThrough(std.NewCodeRealm(gov), func() {
		SetVotingPowerQuorum(0)
	})
	urequire.Equal(t, int64(0), VotingPowerQuorum())

	crossThrough(std.NewCodeRealm(gov), func() {
		SetVotingPowerQuorum(1000)
	})
	urequire.Equal(t, int64(1000), VotingPowerQuorum())
}

func TestProposal_QuorumDisabledEnforcement(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	gov := "gno.land/r/volos/gov/governance"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")

	crossThrough(std.NewCodeRealm(gov), func() {
		SetVotingPowerQuorum(0)
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 1000)
		AddMember(cross, alice)
	})

	var proposalId uint64
	crossThrough(std.NewUserRealm(alice), func() {
		proposalId = CreateProposal(cross, "Disabled Quorum Test", "Body", time.Second*10, func() {}).ID()
	})

	crossThrough(std.NewUserRealm(alice), func() {
		Vote(cross, proposalId, "YES", "For!")
	})

	proposal := volosGovernance.GetProposal(proposalId)
	pass, err := proposal.Definition().(VolosProposalDefinition).Tally(proposal.VotingRecord().Readonly(), MemberSet())
	urequire.NoError(t, err)
	urequire.True(t, pass, "proposal should pass when voting power quorum is disabled (set to 0)")

	crossThrough(std.NewCodeRealm(gov), func() {
		SetVotingPowerQuorum(1000)
	})
}

func TestProposal_NegativeQuorumValues(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	gov := "gno.land/r/volos/gov/governance"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")

	crossThrough(std.NewCodeRealm(gov), func() {
		SetVotingPowerQuorum(-1)
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 1000)
		AddMember(cross, alice)
	})

	var proposalId uint64
	crossThrough(std.NewUserRealm(alice), func() {
		proposalId = CreateProposal(cross, "Negative Quorum Test", "Body", time.Second*10, func() {}).ID()
	})

	crossThrough(std.NewUserRealm(alice), func() {
		Vote(cross, proposalId, "YES", "For!")
	})

	proposal := volosGovernance.GetProposal(proposalId)
	pass, err := proposal.Definition().(VolosProposalDefinition).Tally(proposal.VotingRecord().Readonly(), MemberSet())
	urequire.NoError(t, err)
	urequire.True(t, pass, "proposal should pass when voting power quorum is set to negative values")

	crossThrough(std.NewCodeRealm(gov), func() {
		SetVotingPowerQuorum(1000)
	})
}

func TestProposal_ExecutionLifecycle(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 10000)
		AddMember(cross, alice)
	})

	var executed bool
	var proposal *commondao.Proposal
	crossThrough(std.NewUserRealm(alice), func() {
		proposal = CreateProposal(cross, "Exec Test", "Body", time.Second*3600, func() {
			executed = true
		})
		Vote(cross, proposal.ID(), "YES", "For!")
	})

	crossThrough(std.NewUserRealm(alice), func() {
		err := volosGovernance.Execute(proposal.ID())
		urequire.Error(t, err)
	})

	testing.SkipHeights(2000)
	crossThrough(std.NewUserRealm(alice), func() {
		err := volosGovernance.Execute(proposal.ID())
		urequire.NoError(t, err)
	})
	urequire.True(t, executed)
	proposal = volosGovernance.GetProposal(proposal.ID())
	urequire.Equal(t, "passed", string(proposal.Status()))
}

func TestProposal_QuorumChangesDontAffectActiveProposals(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	gov := "gno.land/r/volos/gov/governance"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")
	bob := std.DerivePkgAddr("gno.land/r/volos/gov/bob")

	crossThrough(std.NewCodeRealm(gov), func() {
		SetProposalThreshold(100)
		SetVotingPowerQuorum(1000)
	})

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Burn(cross, alice, xvls.BalanceOf(alice))
		xvls.Burn(cross, bob, xvls.BalanceOf(bob))
		xvls.Mint(cross, alice, 2000)
		xvls.Mint(cross, bob, 2000)
		AddMember(cross, alice)
		AddMember(cross, bob)
	})

	var proposalId uint64
	crossThrough(std.NewUserRealm(alice), func() {
		proposalId = CreateProposal(cross, "Quorum Change Test", "Body", time.Second*10, func() {}).ID()
	})

	crossThrough(std.NewUserRealm(alice), func() {
		Vote(cross, proposalId, "YES", "For!")
	})
	crossThrough(std.NewUserRealm(bob), func() {
		Vote(cross, proposalId, "YES", "For!")
	})

	proposal := volosGovernance.GetProposal(proposalId)
	pass, err := proposal.Definition().(VolosProposalDefinition).Tally(proposal.VotingRecord().Readonly(), MemberSet())
	urequire.NoError(t, err)
	urequire.True(t, pass, "proposal should pass with original quorum (4000 > 1000)")

	crossThrough(std.NewCodeRealm(gov), func() {
		SetVotingPowerQuorum(10000)
	})

	pass, err = proposal.Definition().(VolosProposalDefinition).Tally(proposal.VotingRecord().Readonly(), MemberSet())
	urequire.NoError(t, err)
	urequire.True(t, pass, "proposal should still pass even after quorum change (uses original quorum)")

	var newProposalId uint64
	crossThrough(std.NewUserRealm(alice), func() {
		newProposalId = CreateProposal(cross, "New Quorum Test", "Body", time.Second*10, func() {}).ID()
	})

	crossThrough(std.NewUserRealm(alice), func() {
		Vote(cross, newProposalId, "YES", "For!")
	})
	crossThrough(std.NewUserRealm(bob), func() {
		Vote(cross, newProposalId, "YES", "For!")
	})

	newProposal := volosGovernance.GetProposal(newProposalId)
	pass, err = newProposal.Definition().(VolosProposalDefinition).Tally(newProposal.VotingRecord().Readonly(), MemberSet())
	urequire.NoError(t, err)
	urequire.False(t, pass, "new proposal should fail with new quorum (4000 < 10000)")

	crossThrough(std.NewCodeRealm(gov), func() {
		SetVotingPowerQuorum(1000)
	})
}

func TestProposal_MaximumDurationEnforced(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	gov := "gno.land/r/volos/gov/governance"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 10_000)
		AddMember(cross, alice)
	})

	max := MaximumProposalDuration()
	urequire.True(t, max > 0)

	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "voting period exceeds maximum allowed duration", func() {
			CreateProposal(cross, "Too long", "Body", max+time.Second, func() {})
		})
	})

	crossThrough(std.NewCodeRealm(gov), func() {
		SetMaximumProposalDuration(max + time.Hour)
	})

	urequire.Equal(t, int64(max+time.Hour), int64(MaximumProposalDuration()))

	crossThrough(std.NewUserRealm(alice), func() {
		p := CreateProposal(cross, "Within new max", "Body", max+time.Second, func() {})
		urequire.NotEqual(t, nil, p)
	})
}
