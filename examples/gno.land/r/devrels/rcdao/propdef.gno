package rcdao

import (
	"std"
	"time"

	"gno.land/p/nt/commondao"
)

type EventProposal struct {
	title           string
	body            string
	location        string // ie Berlin, Germany
	requestedBudget std.Coins
	beneficiary     std.Address // after the prop passes, coins are sent here
}

func (ep EventProposal) Title() string            { return ep.title }
func (ep EventProposal) Body() string             { return ep.body }
func (EventProposal) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

func (EventProposal) Tally(commondao.ReadonlyVotingRecord, commondao.MemberSet) (bool, error) {
	// how do i tally? are there any prebuilt methods?
	return true, nil
}

func (EventProposal) Execute() error {
	crossing()

	return nil
}
