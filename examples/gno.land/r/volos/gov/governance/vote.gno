// Package governance/vote provides voting logic for Volos DAO proposals.
//
// Voting power is determined by the voter's xVLS balance at the time of voting.
package governance

import (
	"std"

	"gno.land/p/nt/commondao"
	"gno.land/r/volos/gov/xvls"
)

// Vote submits a weighted vote for a proposal, using xVLS balance as voting power.
// This functions matches the commondao.Vote() almost completely except for the context which is the xVLS balance of the voter.
func Vote(cur realm, proposalID uint64, choice string, reason string) {
	voter := std.PreviousRealm().Address()
	if !volosGovernance.Members().Has(voter) {
		panic(commondao.ErrNotMember)
	}

	p := volosGovernance.GetProposal(proposalID)
	if p == nil {
		panic(commondao.ErrProposalNotFound)
	}

	if !p.IsVoteChoiceValid(commondao.VoteChoice(choice)) {
		panic(commondao.ErrInvalidVoteChoice)
	}

	vote := commondao.Vote{
		Address: voter,
		Choice:  commondao.VoteChoice(choice),
		Reason:  reason,
		Context: xvls.BalanceOf(voter),
	}
	p.VotingRecord().AddVote(vote)
}
