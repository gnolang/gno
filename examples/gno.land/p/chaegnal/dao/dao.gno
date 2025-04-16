package dao

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
)

const (
	adminRank  = "admin"
	memberRank = "member"

	whitelistFunc    = "WhitelistingAwaitingApproval"
	refuseMemberFunc = "RefuseAwaitingMember"
	setAdminFunc     = "SetAdmin"
	removeAdminFunc  = "RemoveAdminRank"
	removeMemberFunc = "RemoveMember"

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

type DAO struct {
	Proposals        *avl.Tree
	Whitelist        *avl.Tree
	AwaitingApproval *avl.Tree
}

// Create a new DAO
// Initializes proposal, whitelist and waitlist
func NewDAO(caller std.Address) *DAO {
	newDao := &DAO{
		Proposals:        avl.NewTree(),
		Whitelist:        avl.NewTree(),
		AwaitingApproval: avl.NewTree(),
	}

	newDao.WhitelistMember(caller)
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
	isAdmin := false

	member, isMember := dao.IsMember(adr)
	if !isMember {
		return false
	}

	if member.Rank == adminRank {
		isAdmin = true
	}

	return isAdmin
}

// Verify if a user is in the waitlist
func (dao *DAO) IsPending(adr std.Address) bool {
	_, ok := dao.AwaitingApproval.Get(adr.String())
	if !ok {
		return false
	}
	return true
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

// Getters
func (dao *DAO) GetWhitelisted() *avl.Tree {
	return dao.Whitelist
}

func (dao *DAO) GetAwaiting() *avl.Tree {
	return dao.AwaitingApproval
}

func (dao *DAO) GetProposals() *avl.Tree {
	return dao.Proposals
}

// Set user into the waitlist
// Verify if the user is already in the waitlist or in the whitelist
func (dao *DAO) PendingApproval(adr std.Address) {
	_, err := dao.AwaitingApproval.Get(adr.String())
	if err {
		panic("Already awaiting approval.")
	}

	_, err = dao.IsMember(adr)
	if err {
		panic("Already whitelisted.")
	}

	dao.AwaitingApproval.Set(adr.String(), true)
}

// Move user from the waitlist to the whitelist
// Verify that the action is done by an admin
// Verify that the user is in the waitlist and not already in the whitelist
func (dao *DAO) WhitelistingAwaitingApproval(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("Only admins can whitelist members.")
	}

	_, ok := dao.AwaitingApproval.Get(adr.String())
	if !ok {
		panic("This address is not awaiting approval.")
	}

	_, isMember := dao.IsMember(adr)
	if isMember {
		panic("Already whitelisted.")
	}

	dao.WhitelistMember(adr)
	dao.AwaitingApproval.Remove(adr.String())
}

// Remove user from the waitlist
// Verify that the action is done by an admin
func (dao *DAO) RefuseAwaitingMember(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("Only admins can refuse members.")
	}

	dao.AwaitingApproval.Remove(adr.String())
}

// Set a user into the whitelist with a default member rank
func (dao *DAO) WhitelistMember(adr std.Address) {
	newWhitelisted := WhitelistedUser{
		Address: adr,
		Rank:    memberRank,
	}

	dao.Whitelist.Set(adr.String(), newWhitelisted)
}

// Allow admin to set a user as an admin
func (dao *DAO) SetAdmin(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("Only admins can promote members.")
	}

	dao.defaultAdmin(adr)
}

// Set a default admin at the creation of the DAO
// Private function to maximize security
func (dao *DAO) defaultAdmin(adr std.Address) {
	member, ok := dao.IsMember(adr)
	if dao.IsAdmin(adr) {
		panic("Member is already an admin.")
	}

	if !ok {
		panic("Not a member.")
	}

	member.Rank = adminRank
	dao.Whitelist.Set(adr.String(), member)
}

// Demote admin to member
// Verify is the action is made by an admin, if the member exist and is a admin
// Bans auto-demotion
func (dao *DAO) RemoveAdminRank(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("Only admins can demote members.")
	}

	member, ok := dao.IsMember(adr)
	if !dao.IsAdmin(adr) {
		panic("Member is not an admin.")
	}

	if !ok {
		panic("Not a member.")
	}

	if caller == adr {
		panic("Cannot demote yourself.")
	}

	member.Rank = memberRank
	dao.Whitelist.Set(adr.String(), member)
}

// Remove member from whitelist
// Can only be done by an admin on an existing member
// Bans auto-removal
func (dao *DAO) RemoveMember(caller, adr std.Address) {
	if !dao.IsAdmin(caller) {
		panic("Must be an admin to remove a member.")
	}

	member, ok := dao.IsMember(adr)
	if !ok {
		panic("Not a member.")
	}

	if member.Address == caller {
		panic("You cannot remove yourself.")
	}

	dao.Whitelist.Remove(adr.String())
}

// Create new proposal
// Can only be done by an admin
func (dao *DAO) CreateProposals(title, desc, callback string, callbackArgs []string, caller std.Address) {
	if !dao.IsAdmin(caller) {
		panic("You cannot create proposals")
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

// Execute a function when proposal closes if a callback was set
// Verify the number of arguments
func (dao *DAO) ExecuteCallback(proposal Proposal) {
	if proposal.Callback == "" {
		return
	}

	if len(proposal.CallbackArgs) < 2 {
		panic("Not enough arguments.")
	}

	switch proposal.Callback {
	case whitelistFunc:
		dao.WhitelistingAwaitingApproval(std.Address(proposal.CallbackArgs[0]), std.Address(proposal.CallbackArgs[1]))

	case refuseMemberFunc:
		dao.RefuseAwaitingMember(std.Address(proposal.CallbackArgs[0]), std.Address(proposal.CallbackArgs[1]))

	case setAdminFunc:
		dao.SetAdmin(std.Address(proposal.CallbackArgs[0]), std.Address(proposal.CallbackArgs[1]))

	case removeAdminFunc:
		dao.RemoveAdminRank(std.Address(proposal.CallbackArgs[0]), std.Address(proposal.CallbackArgs[1]))

	case removeMemberFunc:
		dao.RemoveMember(std.Address(proposal.CallbackArgs[0]), std.Address(proposal.CallbackArgs[1]))

	default:
		panic("Unknown callback: " + proposal.Callback)
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
		panic("You are not whitelisted.")
	}

	raw, ok := dao.Proposals.Get(strconv.Itoa(voteId))
	if !ok {
		panic("Proposal does not exist.")
	}

	vote := raw.(Proposal)

	if !vote.Active {
		panic("Proposal already closed.")
	}

	_, voted := vote.Voters.Get(caller.String())
	if voted {
		panic("Already voted.")
	}

	if choiceIdx != VoteYay && choiceIdx != VoteNay {
		panic("Invalid vote. Must be 0 (Nay) or 1 (Yay).")
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
		panic("Proposal does not exist.")
	}

	closingProp := raw.(Proposal)
	if !closingProp.Active {
		panic("Proposal already closed.")
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
		panic("You cannot close a proposal.")
	}

	dao.closeProposal(id, caller)
}

// Let member leave the group
// Bans admin from leaving if no other admin exists
func (dao *DAO) LeaveDao(caller std.Address) {
	var adminCounter int

	_, ok := dao.IsMember(caller)
	if !ok {
		panic("Not a member.")
	}

	if dao.IsAdmin(caller) {
		adminCounter = dao.countAdmins()
		if adminCounter <= 1 {
			panic("You are the last admin. Assign another admin before leaving.")
		}
	}

	dao.Whitelist.Remove(caller.String())
}
