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
