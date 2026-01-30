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

const (
	acceptMemberFunc = "acceptAwaitingMember"
	refuseMemberFunc = "refuseAwaitingMember"
	makeAdminFunc    = "makeAdmin"
	removeAdminFunc  = "removeAdminRank"
	removeMemberFunc = "removeMember"
	setHangoutFunc   = "SetHangoutTimeAndPlace"
)

type callbackProposalDefinition struct {
	GroupId      string
	TitleValue   string
	BodyValue    string
	Callback     string
	CallbackArgs []string
}

func (c *callbackProposalDefinition) Title() string {
	return c.TitleValue
}

func (c *callbackProposalDefinition) Body() string {
	return c.BodyValue
}

func (c *callbackProposalDefinition) VotingPeriod() time.Duration {
	return 24 * time.Hour
}

func (c *callbackProposalDefinition) Tally(r commondao.ReadonlyVotingRecord, m commondao.MemberSet) (bool, error) {
	yesCount := r.VoteCount(commondao.ChoiceYes)
	noCount := r.VoteCount(commondao.ChoiceNo)
	return yesCount > noCount, nil
}

// CreateProposals allows an admin to create a proposal
// It sanitizes the arguments given
func CreateProposals(cur realm, groupId, title, desc, callback, callbackArgsStr, logged string) {
	if !isAdmin(groupId, std.PreviousRealm().Address()) {
		panic("only admin can create proposals")
	}

	group := organizer.GetGroupFromId(groupId)

	callbackArgsStr = strings.Trim(callbackArgsStr, "\"")
	callbackArgs := strings.Split(callbackArgsStr, ",")

	def := &callbackProposalDefinition{
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

// VoteOnProposal allows the user to vote on a proposal
// Checks if the proposal deadline is overdue
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

// Execute defines the authorized callbacks for proposals
// Execute start the execution of the desired callback
func (p *callbackProposalDefinition) Execute(cur realm) error {
	group := organizer.GetGroupFromId(p.GroupId)

	if len(p.CallbackArgs) < 2 {
		return errors.New("not enough arguments")
	}

	switch p.Callback {
	case acceptMemberFunc:
		return acceptAwaitingMember(cross, p.GroupId, p.CallbackArgs[0], std.Address(p.CallbackArgs[1]))

	case refuseMemberFunc:
		return refuseAwaitingMember(cross, p.GroupId, p.CallbackArgs[0], std.Address(p.CallbackArgs[1]))

	case makeAdminFunc:
		return makeAdmin(cross, p.GroupId, p.CallbackArgs[0], std.Address(p.CallbackArgs[1]))

	case removeAdminFunc:
		return removeAdminRank(cross, p.GroupId, p.CallbackArgs[0], std.Address(p.CallbackArgs[1]))

	case removeMemberFunc:
		return removeMember(cross, p.GroupId, p.CallbackArgs[0], std.Address(p.CallbackArgs[1]))

	case setHangoutFunc:
		if len(p.CallbackArgs) < 4 {
			return errors.New("not enough arguments")
		}
		return group.SetHangoutTimeAndPlace(p.CallbackArgs[0], p.CallbackArgs[1], p.CallbackArgs[2], p.CallbackArgs[3])

	default:
		return errors.New("unknown callback: " + p.Callback)
	}
}

// ExecuteExpiredProposals checks if the voting deadline is overdue
// Calls the function to execute the callback if so
func ExecuteExpiredProposals(cur realm, groupId string) {
	group := organizer.GetGroupFromId(groupId)

	group.DAO.ActiveProposals().Iterate(0, group.DAO.ActiveProposals().Size(), false, func(p *commondao.Proposal) bool {
		if time.Now().After(p.VotingDeadline()) {
			_ = group.DAO.Execute(p.ID())
		}
		return true
	})

}
