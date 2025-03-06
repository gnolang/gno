package minisocial

import (
	"github.com/gnolang/gno/gnovm/stdlibs/std"
	"gno.land/p/nt/commondao"
	cmndao "gno.land/r/nt/commondao"
)

var DAO = cmndao.New("My DAO")

func init() {
	DAO.Members().Add("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn
}

// RemovePostProp implements ProposalDefinition
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

// what if i want some specific context that I
// hardcode at the time of the proposal? ie:
// i want to propose a specific balance/budget exactly at the price when
// i propose something... -> Take a look at govdao v3 "Metadata"
// the issue is that Execute will be called at some point,
// with some state, which has advanced since I made the proposal
// I should be able to do either:
// execute prop at time when its voted yes (ie coin*amt == 10$)
// or, record price at time of exec, get val of coin, send exactly 10$
// at the time of execution of the prop

/// sidenote: please review govdao v3!!! super important :)

func (p RemovePostProp) Execute() error {
	_, removed := posts.Remove(p.postToRemoveID)
	if !removed {
		panic("Removal failed")
	}

	return nil
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

func (RemovePostProp) Quorum() float64 { return 0.67 } // two thirds

// why do i need to define the argments?
// feels like this should't be exported, who will call this just to get hte tally?
// I want a prop.Tally to return either "not voted on yet", or
// if you need to defne a tally then
// Tally how do i as a user define this readonly votingt record?
func (p RemovePostProp) Tally(r commondao.ReadOnlyVotingRecord, membersCount int) bool {
	_, success := commondao.SelectChoiceByAbsoluteMajority(r, membersCount)
	return success
}
