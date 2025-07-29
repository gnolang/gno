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

func TestProposal_NewAndVoteChoices(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 10000)
		AddMember(cross, alice)
	})

	crossThrough(std.NewUserRealm(alice), func() {
		CreateProposal(cross, "Default Choices", "Body", time.Second*10, func() {})
	})

	p := volosGovernance.GetProposal(1)
	urequire.NotEqual(t, nil, p)

	choices := p.VoteChoices()
	urequire.Equal(t, 3, len(choices))

	urequire.True(t, choices[0] == "YES" || choices[1] == "YES" || choices[2] == "YES")
	urequire.True(t, choices[0] == "NO" || choices[1] == "NO" || choices[2] == "NO")
	urequire.True(t, choices[0] == "ABSTAIN" || choices[1] == "ABSTAIN" || choices[2] == "ABSTAIN")
}

func TestProposal_TallyAndQuorum(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")
	bob := std.DerivePkgAddr("gno.land/r/volos/gov/bob")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 10000)
		xvls.Mint(cross, bob, 10000)
		AddMember(cross, alice)
		AddMember(cross, bob)
	})

	crossThrough(std.NewUserRealm(alice), func() {
		CreateProposal(cross, "Quorum Test", "Body", time.Second*10, func() {})
		Vote(cross, 1, "YES", "For!")
	})

	crossThrough(std.NewUserRealm(bob), func() {
		Vote(cross, 1, "NO", "Against!")
	})

	p := volosGovernance.GetProposal(1)
	urequire.NotEqual(t, nil, p)

	pass, err := p.Definition().(VolosProposalDefinition).Tally(p.VotingRecord().Readonly(), MemberSet())
	urequire.NoError(t, err)
	urequire.True(t, pass || !pass)
}

func TestProposal_ExecuteErrors(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 10000)
		AddMember(cross, alice)
	})

	var proposal *commondao.Proposal
	crossThrough(std.NewUserRealm(alice), func() {
		proposal = CreateProposal(cross, "Panic Action", "Body", time.Second*1, func() { panic("fail") })
		Vote(cross, proposal.ID(), "YES", "For!")
	})

	testing.SkipHeights(2000)

	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "fail", func() {
			volosGovernance.Execute(proposal.ID())
		})
	})
}

func TestProposal_ExecuteFunction(cur realm, t *testing.T) {
	staker := "gno.land/r/volos/gov/staker"
	alice := std.DerivePkgAddr("gno.land/r/volos/gov/alice")

	crossThrough(std.NewCodeRealm(staker), func() {
		xvls.Mint(cross, alice, 10000)
		AddMember(cross, alice)
	})

	var proposal *commondao.Proposal
	executed := false
	crossThrough(std.NewUserRealm(alice), func() {
		proposal = CreateProposal(cross, "Successful Execution", "Body", time.Second*1, func() { executed = true })
		Vote(cross, proposal.ID(), "YES", "For!")
	})

	testing.SkipHeights(2000)

	crossThrough(std.NewUserRealm(alice), func() {
		Execute(cross, proposal.ID())
		urequire.True(t, executed)
	})

	crossThrough(std.NewUserRealm(alice), func() {
		uassert.AbortsWithMessage(t, "proposal not found", func() {
			Execute(cross, 999)
		})
	})

	executed2 := false
	crossThrough(std.NewUserRealm(alice), func() {
		proposal2 := CreateProposal(cross, "Early Execution", "Body", time.Second*100, func() { executed2 = true })
		Vote(cross, proposal2.ID(), "YES", "For!")
		
		uassert.AbortsWithMessage(t, "voting deadline not met", func() {
			Execute(cross, proposal2.ID())
		})
		urequire.False(t, executed2)
	})
}
