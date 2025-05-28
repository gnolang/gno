package group

import (
	"std"
	"strconv"
	"strings"

	"gno.land/r/chaegnal/organizer"
)

// CreateProposals is a wrapper to create a proposal
// It sanitizes the arguments given
func CreateProposals(groupId, title, desc, callback, callbackArgsStr, logged string) {
	crossing()
	group := organizer.GetGroupFromId(groupId)

	callbackArgsStr = strings.Trim(callbackArgsStr, "\"")
	callbackArgs := strings.Split(callbackArgsStr, ",")

	group.DAO.CreateProposals(title, desc, callback, callbackArgs, std.Address(logged))
}

// VoteOnProposal is a wrapper to vote on a proposal
// Panics if the given vote informations are not possible
func VoteOnProposal(groupId, voteIdStr, voteStr, logged string) {
	crossing()
	group := organizer.GetGroupFromId(groupId)

	voteId, err := strconv.Atoi(voteIdStr)
	if err != nil {
		panic("invalid vote id")
	}

	vote, err := strconv.Atoi(voteStr)
	if err != nil {
		panic("invalid vote choice")
	}

	if vote != organizer.VoteNay && vote != organizer.VoteYay {
		panic("vote should be yay (1) or nay (0)")
	}

	group.DAO.VoteProposal(voteId, vote, std.Address(logged))
}

// AdminCloseProposal is a wrapper for admin to close proposal forcefully
// Panics if the vote ID does not belong to any open proposals
func AdminCloseProposal(groupId, voteIdStr, logged string) {
	crossing()
	group := organizer.GetGroupFromId(groupId)

	voteId, err := strconv.Atoi(voteIdStr)
	if err != nil {
		panic("invalid vote id")
	}

	group.DAO.AdminCloseProposal(voteId, std.Address(logged))
}
