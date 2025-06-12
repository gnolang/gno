package group

import (
	"errors"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/nt/commondao"
	"gno.land/r/chaegnal/organizer"
)

type CallbackProposalDefinition struct {
	GroupId      string
	TitleValue   string
	BodyValue    string
	Callback     string
	CallbackArgs []string
}

var _ commondao.Executable = (*CallbackProposalDefinition)(nil)

type ProposalDefinition interface {
	Title() string
	Body() string
	VotingPeriod() time.Duration
	Tally(commondao.ReadonlyVotingRecord, commondao.MemberSet) (bool, error)
}

func (c *CallbackProposalDefinition) Title() string {
	return c.TitleValue
}

func (c *CallbackProposalDefinition) Body() string {
	return c.BodyValue
}

func (c *CallbackProposalDefinition) VotingPeriod() time.Duration {
	return 1 * time.Minute
}

func (c *CallbackProposalDefinition) Tally(r commondao.ReadonlyVotingRecord, m commondao.MemberSet) (bool, error) {
	yesCount := r.VoteCount(commondao.ChoiceYes)
	noCount := r.VoteCount(commondao.ChoiceNo)
	return yesCount > noCount, nil
}

// CreateProposals is a wrapper to create a proposal
// It sanitizes the arguments given
func CreateProposals(cur realm, groupId, title, desc, callback, callbackArgsStr, logged string) {
	if !isAdmin(groupId, std.PreviousRealm().Address()) {
		panic("only admin can create proposals")
	}

	group := organizer.GetGroupFromId(groupId)

	callbackArgsStr = strings.Trim(callbackArgsStr, "\"")
	callbackArgs := strings.Split(callbackArgsStr, ",")

	def := &CallbackProposalDefinition{
		GroupId:      groupId,
		TitleValue:   title,
		BodyValue:    desc,
		Callback:     callback,
		CallbackArgs: callbackArgs,
	}

	_, err := group.DAO.Propose(std.Address(logged), def)
	if err != nil {
		panic("proposal creation failed: " + err.Error())
	}
}

func VoteOnProposal(cur realm, groupId, voteIdStr, voteStr, logged string) {
	group := organizer.GetGroupFromId(groupId)

	voteId, err := strconv.ParseUint(voteIdStr, 10, 64)
	if err != nil {
		panic("invalid vote id")
	}

	voteInt, err := strconv.Atoi(voteStr)
	if err != nil {
		panic("invalid vote choice")
	}

	var choice commondao.VoteChoice
	switch voteInt {
	case organizer.VoteYay:
		choice = commondao.ChoiceYes
	case organizer.VoteNay:
		choice = commondao.ChoiceNo
	default:
		panic("vote should be yay (1) or nay (0)")
	}

	err = group.DAO.Vote(std.Address(logged), voteId, choice, "")
	if err != nil {
		panic("failed to vote: " + err.Error())
	}
	ExecuteExpiredProposals(cross, groupId)
}

// AdminCloseProposal is a wrapper for admin to close proposal forcefully
// Panics if the vote ID does not belong to any open proposals
func AdminCloseProposal(groupId, voteIdStr, logged string) {
	// crossing()
	// group := organizer.GetGroupFromId(groupId)
	//
	// voteId, err := strconv.ParseUint(voteIdStr, 10, 64)
	// if err != nil {
	// panic("invalid vote id")
	// }
	//
	// p := group.DAO.ActiveProposals().Get(voteId)
	// p.SetStatus(commondao.StatusFailed)
	// p.SetStatusReason("forcefully closed by admin")
	// if err != nil {
	// group.DAO.ActiveProposals().Remove(p.ID())
	// group.DAO.FinishedProposals().Add(p)
	// } else {
	// panic("proposal not active")
	// }
}

func (p *CallbackProposalDefinition) Execute() error {
	switch p.Callback {
	case "SetHangout":
		if len(p.CallbackArgs) < 3 {
			return errors.New("not enough arguments")
		}

		group := organizer.GetGroupFromId(p.GroupId)
		return group.SetHangoutTimeAndPlace(p.CallbackArgs[0], p.CallbackArgs[1], p.CallbackArgs[2], p.CallbackArgs[3])

	default:
		return errors.New("unknown callback: " + p.Callback)
	}
}

func ExecuteExpiredProposals(cur realm, groupId string) {
	group := organizer.GetGroupFromId(groupId)

	group.DAO.ActiveProposals().Iterate(0, group.DAO.ActiveProposals().Size(), false, func(p *commondao.Proposal) bool {
		if time.Now().After(p.VotingDeadline()) {
			_ = group.DAO.Execute(p.ID())
		}
		return true
	})

}
