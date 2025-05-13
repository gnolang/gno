package dao

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
)

type Proposal struct {
	Id           int
	Title        string
	Description  string
	Owner        std.Address
	Callback     string
	CallbackArgs []string
	Votes        *avl.Tree
	Voters       *avl.Tree
	Active       bool
	Winner       string
}

type HangoutInfo struct {
	Id          int
	Date        string
	Time        string
	Place       string
	Description string
	Active      bool
}

const (
	whitelistFunc    = "WhitelistUser"
	refuseMemberFunc = "RefuseAwaitingMember"
	setAdminFunc     = "SetAdmin"
	removeAdminFunc  = "RemoveAdminRank"
	removeMemberFunc = "RemoveMember"
	setHangoutFunc   = "SetHangoutTimeAndPlace"

	VoteYay = 1
	VoteNay = 0
)

// Create new proposal
// Can only be done by an admin
func (dao *DAO) CreateProposals(title, desc, callback string, callbackArgs []string, caller std.Address) {
	if !dao.IsAdmin(caller) {
		panic("you cannot create proposals")
	}

	newProposal := Proposal{
		Id:           dao.Proposals.Size() + 1,
		Title:        title,
		Description:  desc,
		Owner:        caller,
		Callback:     callback,
		CallbackArgs: callbackArgs,
		Votes:        avl.NewTree(),
		Voters:       avl.NewTree(),
		Active:       true,
	}

	dao.Proposals.Set(strconv.Itoa(newProposal.Id), newProposal)
}

// Getters
func (dao *DAO) GetProposals() *avl.Tree {
	return dao.Proposals
}

func (dao *DAO) GetHangouts() rotree.IReadOnlyTree {
	return rotree.Wrap(dao.Hangouts, nil)
}

// Let whitelisted member vote on proposals
// Verify the sanity of proposal and voting choice
// Close proposal if every member has voted and execute callback
func (dao *DAO) VoteProposal(voteId, choiceIdx int, caller std.Address) {
	_, ok := dao.IsMember(caller)
	if !ok {
		panic("you are not whitelisted")
	}

	raw, ok := dao.Proposals.Get(strconv.Itoa(voteId))
	if !ok {
		panic("proposal does not exist")
	}

	vote := raw.(Proposal)

	if !vote.Active {
		panic("proposal already closed")
	}

	_, voted := vote.Voters.Get(caller.String())
	if voted {
		panic("already voted")
	}

	if choiceIdx != VoteYay && choiceIdx != VoteNay {
		panic("invalid vote: must be 0 (nay) or 1 (yay)")
	}

	rawVote, _ := vote.Votes.Get(strconv.Itoa(choiceIdx))
	currentVote := 0
	if rawVote != nil {
		currentVote = rawVote.(int)
	}

	vote.Votes.Set(strconv.Itoa(choiceIdx), currentVote+1)
	vote.Voters.Set(caller.String(), choiceIdx)

	dao.Proposals.Set(strconv.Itoa(voteId), vote)

	if dao.hasEveryoneVoted(vote) {
		dao.closeProposal(voteId, caller)
	}
}

// Set Hangout in a avl.tree
func (dao *DAO) SetHangoutTimeAndPlace(date, time, place, desc string, caller std.Address) {
	if !dao.IsAdmin(caller) {
		panic("you cannot create proposals")
	}

	hangout := HangoutInfo{
		Id:          dao.Hangouts.Size() + 1,
		Date:        date,
		Time:        time,
		Place:       place,
		Description: desc,
		Active:      true,
	}

	dao.Hangouts.Set(strconv.Itoa(dao.Hangouts.Size()+1), hangout)
}

// Execute a function when proposal closes if a callback was set
// Verify the number of arguments
func (dao *DAO) ExecuteCallback(proposal Proposal) {
	if proposal.Callback == "" {
		return
	}

	if len(proposal.CallbackArgs) < 2 {
		panic("not enough arguments")
	}

	switch proposal.Callback {
	case whitelistFunc:
		dao.WhitelistUser(std.Address(proposal.CallbackArgs[0]), std.Address(proposal.CallbackArgs[1]))

	case refuseMemberFunc:
		dao.RefuseAwaitingMember(std.Address(proposal.CallbackArgs[0]), std.Address(proposal.CallbackArgs[1]))

	case setAdminFunc:
		dao.SetAdmin(std.Address(proposal.CallbackArgs[0]), std.Address(proposal.CallbackArgs[1]))

	case removeAdminFunc:
		dao.RemoveAdminRank(std.Address(proposal.CallbackArgs[0]), std.Address(proposal.CallbackArgs[1]))

	case removeMemberFunc:
		dao.RemoveMember(std.Address(proposal.CallbackArgs[0]), std.Address(proposal.CallbackArgs[1]))

	case setHangoutFunc:
		if len(proposal.CallbackArgs) < 5 {
			panic("not enough arguments for SetHangoutTimeAndPlace")
		}
		dao.SetHangoutTimeAndPlace(proposal.CallbackArgs[0], proposal.CallbackArgs[1],
			proposal.CallbackArgs[2], proposal.CallbackArgs[3], std.Address(proposal.CallbackArgs[4]))

	default:
		panic("unknown callback: " + proposal.Callback)
	}
}

// Deactivate proposal after verifying sanity
// Sets the winner, default to refusal if no clear winner
// Private member as no admin check
func (dao *DAO) closeProposal(id int, caller std.Address) {
	raw, ok := dao.Proposals.Get(strconv.Itoa(id))
	if !ok {
		panic("proposal does not exist")
	}

	closingProp := raw.(Proposal)
	if !closingProp.Active {
		panic("proposal already closed")
	}

	yayVotesRaw, _ := closingProp.Votes.Get(strconv.Itoa(VoteYay))
	nayVotesRaw, _ := closingProp.Votes.Get(strconv.Itoa(VoteNay))

	yayVotes := 0
	if yayVotesRaw != nil {
		yayVotes = yayVotesRaw.(int)
	}
	nayVotes := 0
	if nayVotesRaw != nil {
		nayVotes = nayVotesRaw.(int)
	}

	if yayVotes > nayVotes {
		closingProp.Winner = "yay"
		dao.ExecuteCallback(closingProp)
	} else {
		closingProp.Winner = "nay"
	}

	closingProp.Active = false
	dao.Proposals.Set(strconv.Itoa(id), closingProp)
}

// Let admin close a proposal prematurely
func (dao *DAO) AdminCloseProposal(id int, caller std.Address) {
	ok := dao.IsAdmin(caller)
	if !ok {
		panic("you cannot close a proposal")
	}

	dao.closeProposal(id, caller)
}

// Helpers

// Verify if every whitlisted member has voted
func (dao *DAO) hasEveryoneVoted(vote Proposal) bool {
	if vote.Voters.Size() == dao.Whitelist.Size() {
		return true
	}
	return false
}
