package minisocial

import (
	"std"
	"time"

	"gno.land/p/nt/commondao"

	cmndao "gno.land/r/nt/commondao"
)

var DAO = cmndao.New("Minisocial DAO")

func init() {
	DAO.Members().Add(Ownable.Owner())
}

func ProposePostRemoval(id, reason string) error {
	if _, err := DAO.Propose(std.PreviousRealm().Address(), RemovePostProp{
		postToRemoveID: id,
		reason:         reason,
	}); err != nil {
		return err
	}

	return nil
}

// Implements ProposalDefinition
type RemovePostProp struct {
	postToRemoveID string
	reason         string
}

func (RemovePostProp) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

func (p RemovePostProp) Title() string {
	return "Proposal to remove post " + p.postToRemoveID
}

func (p RemovePostProp) Body() string {
	return "We'd like to remove the post because " + p.reason
}

func (p RemovePostProp) Execute() error {
	_, removed := posts.Remove(p.postToRemoveID)
	if !removed {
		panic("Removal failed")
	}

	return nil
}

func (RemovePostProp) Quorum() float64 { return 0.67 } // two thirds

func (p RemovePostProp) Tally(r commondao.ReadOnlyVotingRecord, membersCount int) bool {
	_, success := commondao.SelectChoiceByAbsoluteMajority(r, membersCount)
	return success
}
