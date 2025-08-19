// Package governance/proposal defines proposal creation logic for Volos governance.
//
// To create a proposal, the proposer must hold at least ProposalThreshold xVLS tokens.
// The threshold is a global variable and can be changed by governance (through executing a proposal).
package governance

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/nt/commondao"
	"gno.land/r/volos/gov/xvls"
)

type VolosProposalDefinition struct {
	TitleField        string
	BodyField         string
	VotingPeriodField time.Duration
	Action            func()
	QuorumAtCreation  int64 // Store quorum value at proposal creation time
}

func (v VolosProposalDefinition) Title() string               { return v.TitleField }
func (v VolosProposalDefinition) Body() string                { return v.BodyField }
func (v VolosProposalDefinition) VotingPeriod() time.Duration { return v.VotingPeriodField }

// Tally implements the commondao.ProposalDefinition interface.
// This function is used to tally the votes for the proposal meaning that all the logic for determining the result of the proposal is implemented here.
// The result is calculated based on the voting power of the yes and no vote(r)s, and must meet voting power quorum requirements.
func (v VolosProposalDefinition) Tally(r commondao.ReadonlyVotingRecord, m commondao.MemberSet) (bool, error) {
	votePower := avl.NewTree()
	var totalPower int64 = 0

	m.IterateByOffset(0, m.Size(), func(addr std.Address) bool {
		vote, found := r.GetVote(addr)
		if found && vote.Choice != "" {
			power := int64(0)
			if vote.Context != nil {
				power = vote.Context.(int64)
			}

			key := string(vote.Choice)
			var current int64 = 0
			if v, ok := votePower.Get(key); ok {
				current = v.(int64)
			}

			votePower.Set(key, current+power)
			totalPower += power
		}

		return false
	})

	if v.QuorumAtCreation > 0 && totalPower < v.QuorumAtCreation {
		return false, nil
	}

	yesPower, _ := votePower.Get(string(commondao.ChoiceYes))
	noPower, _ := votePower.Get(string(commondao.ChoiceNo))
	if yesPower == nil {
		yesPower = int64(0)
	}

	if noPower == nil {
		noPower = int64(0)
	}

	if yesPower.(int64) > noPower.(int64) {
		return true, nil
	}

	return false, nil
}

func (v VolosProposalDefinition) Execute(cur realm) error {
	if v.Action != nil {
		v.Action()
	}

	return nil
}

// CreateProposal checks the proposer's xVLS balance and, if sufficient, creates a commondao proposal.
// The minimum xVLS required is set in governance.gno and can be changed via SetProposalThreshold.
func CreateProposal(cur realm, title, body string, votingPeriod time.Duration, action func()) *commondao.Proposal {
	proposer := std.PreviousRealm().Address()
	if xvls.BalanceOf(proposer) < proposalThreshold {
		panic(ErrInsufficientXVLS)
	}

	def := VolosProposalDefinition{
		TitleField:        title,
		BodyField:         body,
		VotingPeriodField: votingPeriod,
		Action:            action,
		QuorumAtCreation:  votingPowerQuorum,
	}

	proposal := volosGovernance.MustPropose(proposer, def)
	deadline := time.Now().Add(votingPeriod).Unix()
	emitProposalCreated(proposer, proposal.ID(), title, body, deadline, votingPowerQuorum)
	return proposal
}

// GetProposal returns the proposal with the given ID from the DAO.
func GetProposal(id uint64) *commondao.Proposal {
	return volosGovernance.GetProposal(id)
}

// Execute executes a proposal with the given ID.
// This function calls the commondao's Execute method and panics if an error occurs.
func Execute(cur realm, proposalID uint64) {
	caller := std.PreviousRealm().Address()
	err := volosGovernance.Execute(proposalID)
	if err != nil {
		panic(err)
	}

	emitProposalExecuted(caller, proposalID)
}

// GetUserActiveProposals returns all active proposals that the given user has voted on.
// It iterates the active proposals and checks the voting record for the user.
func GetUserActiveProposals(user std.Address) []*commondao.Proposal {
	proposals := []*commondao.Proposal{}
	active := volosGovernance.ActiveProposals()
	count := active.Size()
	active.Iterate(0, count, false, func(p *commondao.Proposal) bool {
		if p != nil && p.VotingRecord().Readonly().HasVoted(user) {
			proposals = append(proposals, p)
		}
		return false
	})
	return proposals
}
