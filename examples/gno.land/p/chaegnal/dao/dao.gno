package dao

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
)

const (
	adminRank  = "admin"
	memberRank = "member"

	whitelistFunc    = "WhitelistUser"
	refuseMemberFunc = "RefuseAwaitingMember"
	setAdminFunc     = "SetAdmin"
	removeAdminFunc  = "RemoveAdminRank"
	removeMemberFunc = "RemoveMember"
	setHangoutFunc   = "SetHangoutTimeAndPlace"

	VoteYay = 1
	VoteNay = 0
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

type WhitelistedUser struct {
	Address std.Address
	Rank    string
}

type HangoutInfo struct {
	Id          int
	Date        string
	Time        string
	Place       string
	Description string
	Active      bool
}

type DAO struct {
	Proposals        *avl.Tree
	Whitelist        *avl.Tree
	AwaitingApproval *avl.Tree
	Hangouts         *avl.Tree
}

// Create a new DAO
// Initializes proposal, whitelist and waitlist
func NewDAO(caller std.Address) *DAO {
	newDao := &DAO{
		Proposals:        avl.NewTree(),
		Whitelist:        avl.NewTree(),
		AwaitingApproval: avl.NewTree(),
		Hangouts:         avl.NewTree(),
	}

	newDao.whitelistMember(caller)
	newDao.defaultAdmin(caller)

	return newDao
}

// Verify that a user is in the whitelist
func (dao *DAO) IsMember(adr std.Address) (WhitelistedUser, bool) {
	raw, ok := dao.Whitelist.Get(adr.String())
	if !ok {
		return WhitelistedUser{}, false
	}

	return raw.(WhitelistedUser), true
}

// Verify that a user is in the whitelist and possesses the admin rank
func (dao *DAO) IsAdmin(adr std.Address) bool {
	member, isMember := dao.IsMember(adr)

	return isMember && member.Rank == adminRank
}

// Verify if a user is in the waitlist
func (dao *DAO) IsPending(adr std.Address) bool {
	_, ok := dao.AwaitingApproval.Get(adr.String())

	return ok
}

// Getters
func (dao *DAO) GetWhitelisted() rotree.IReadOnlyTree {
	return rotree.Wrap(dao.Whitelist, nil)
}

func (dao *DAO) GetAwaiting() rotree.IReadOnlyTree {
	return rotree.Wrap(dao.AwaitingApproval, nil)
}

func (dao *DAO) GetProposals() *avl.Tree {
	return dao.Proposals
}

func (dao *DAO) GetHangouts() rotree.IReadOnlyTree {
	return rotree.Wrap(dao.Hangouts, nil)
}

// Set user into the waitlist
// Verify if the user is already in the waitlist or in the whitelist
func (dao *DAO) PendingApproval(adr std.Address) {
	_, err := dao.AwaitingApproval.Get(adr.String())
	if err {
		panic("already awaiting approval")
	}

	_, err = dao.IsMember(adr)
	if err {
		panic("already whitelisted")
	}

	dao.AwaitingApproval.Set(adr.String(), true)
}

// Move user from the waitlist to the whitelist
// Verify that the action is done by an admin
// Verify that the user is in the waitlist and not already in the whitelist
func (dao *DAO) WhitelistUser(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("only admins can whitelist members")
	}

	_, ok := dao.AwaitingApproval.Get(adr.String())
	if !ok {
		panic("this address is not awaiting approval")
	}

	_, isMember := dao.IsMember(adr)
	if isMember {
		panic("already whitelisted")
	}

	dao.whitelistMember(adr)
	dao.AwaitingApproval.Remove(adr.String())
}

// Remove user from the waitlist
// Verify that the action is done by an admin
func (dao *DAO) RefuseAwaitingMember(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("only admins can refuse members")
	}

	dao.AwaitingApproval.Remove(adr.String())
}

// Set a user into the whitelist with a default member rank
func (dao *DAO) whitelistMember(adr std.Address) {
	newWhitelisted := WhitelistedUser{
		Address: adr,
		Rank:    memberRank,
	}

	dao.Whitelist.Set(adr.String(), newWhitelisted)
}

// Allow admin to set a user as an admin
func (dao *DAO) SetAdmin(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("only admins can promote members")
	}

	dao.defaultAdmin(adr)
}

// Set a default admin at the creation of the DAO
// Private function to maximize security
func (dao *DAO) defaultAdmin(adr std.Address) {
	member, ok := dao.IsMember(adr)
	if dao.IsAdmin(adr) {
		panic("member is already an admin")
	}

	if !ok {
		panic("not a member")
	}

	member.Rank = adminRank
	dao.Whitelist.Set(adr.String(), member)
}

// Demote admin to member
// Verify is the action is made by an admin, if the member exist and is a admin
// Bans auto-demotion
func (dao *DAO) RemoveAdminRank(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("only admins can demote members")
	}

	member, _ := dao.IsMember(adr)
	if !dao.IsAdmin(adr) {
		panic("member is not an admin")
	}

	if caller == adr {
		panic("cannot demote yourself")
	}

	member.Rank = memberRank
	dao.Whitelist.Set(adr.String(), member)
}

// Remove member from whitelist
// Can only be done by an admin on an existing member
// Bans auto-removal
func (dao *DAO) RemoveMember(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("must be an admin to remove a member")
	}

	member, ok := dao.IsMember(adr)
	if !ok {
		panic("not a member")
	}

	if member.Address == caller {
		panic("you cannot remove yourself")
	}

	dao.Whitelist.Remove(adr.String())
}

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

// Verify if every whitlisted member has voted
func (dao *DAO) hasEveryoneVoted(vote Proposal) bool {
	if vote.Voters.Size() == dao.Whitelist.Size() {
		return true
	}
	return false
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

// Let member leave the group
// Bans admin from leaving if no other admin exists
func (dao *DAO) LeaveDao(caller std.Address) {
	var adminCounter int

	_, ok := dao.IsMember(caller)
	if !ok {
		panic("not a member")
	}

	if dao.IsAdmin(caller) {
		adminCounter = dao.countAdmins()
		if adminCounter <= 1 {
			panic("you are the last admin: assign another admin before leaving")
		}
	}

	dao.Whitelist.Remove(caller.String())
}

func (dao *DAO) countAdmins() int {
	var adminCounter int

	dao.Whitelist.Iterate("", "", func(_ string, val any) bool {
		user := val.(WhitelistedUser)

		if user.Rank == adminRank {
			adminCounter++
		}

		return false
	})

	return adminCounter
}
