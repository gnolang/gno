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
}

func (v VolosProposalDefinition) Title() string               { return v.TitleField }
func (v VolosProposalDefinition) Body() string                { return v.BodyField }
func (v VolosProposalDefinition) VotingPeriod() time.Duration { return v.VotingPeriodField }

// Tally implements the commondao.ProposalDefinition interface.
// This function is used to tally the votes for the proposal meaning that all the logic for determining the result of the proposal is implemented here.
// The result is calculated based on the voting power of the yes and no vote(r)s.
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

func (v VolosProposalDefinition) Validate() error {
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
	}

	proposal := volosGovernance.MustPropose(proposer, def)
	return proposal
}

// GetProposal returns the proposal with the given ID from the DAO.
func GetProposal(id uint64) *commondao.Proposal {
	return volosGovernance.GetProposal(id)
}
