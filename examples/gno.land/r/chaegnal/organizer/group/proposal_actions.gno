package group

import (
	"std"
	"strconv"
	"strings"

	"gno.land/r/chaegnal/organizer"
)

// Wrapper to create proposal
// Sanitizes the arguments given
func CreateProposals(groupId, title, desc, callback, callbackArgsStr, logged string) {
	group := organizer.GetGroupFromId(groupId)

	callbackArgsStr = strings.Trim(callbackArgsStr, "\"")
	callbackArgs := strings.Split(callbackArgsStr, ",")

	group.DAO.CreateProposals(title, desc, callback, callbackArgs, std.Address(logged))
}

// Wrapper to vote on proposal
// Verify if the vote ID exist and if the vote's choice exists
func VoteOnProposal(groupId, voteIdStr, voteStr, logged string) {
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

// Wrapper for admin to close proposal forcefully
func AdminCloseProposal(groupId, voteIdStr, logged string) {
	group := organizer.GetGroupFromId(groupId)

	voteId, err := strconv.Atoi(voteIdStr)
	if err != nil {
		panic("invalid vote id")
	}

	group.DAO.AdminCloseProposal(voteId, std.Address(logged))
}
