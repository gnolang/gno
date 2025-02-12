package governance

import (
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/json"
)

func createProposalJsonNode(id uint64, proposal ProposalInfo) *json.Node {
	return json.Builder().
		WriteString("id", formatUint(id)).
		WriteString("configVersion", formatUint(proposal.ConfigVersion)).
		WriteString("proposer", proposal.Proposer.String()).
		WriteString("status", b64Encode(getProposalStatus(id))).
		WriteString("type", proposal.ProposalType.String()).
		WriteString("title", proposal.Title).
		WriteString("description", proposal.Description).
		WriteString("vote", b64Encode(getProposalVotes(id))).
		WriteString("extra", b64Encode(getProposalExtraData(id))).
		Node()
}

// GetProposals returns all proposals with necessary information.
func GetProposals() string {
	updateProposalsState()

	if proposals.Size() == 0 {
		return ""
	}

	proposalsObj := metaNode()
	proposalArr := json.ArrayNode("", nil)

	proposals.Iterate("", "", func(key string, value interface{}) bool {
		proposalObj := getProposalById(proposalId)
		proposalArr.AppendArray(proposalObj)
		return false
	})

	proposalsObj.AppendObject("proposals", proposalArr)

	return marshal(proposalsObj)
}

// GetProposalById returns a single proposal with necessary information.
func GetProposalById(id uint64) string {
	updateProposalsState()

	_, exists := proposals.Get(formatUint(id))
	if !exists {
		return ""
	}

	proposalsObj := metaNode()
	proposalArr := json.ArrayNode("", nil)
	proposalObj := getProposalById(id)
	proposalArr.AppendArray(proposalObj)
	proposalsObj.AppendObject("proposals", proposalArr)

	return marshal(proposalsObj)
}

// helper function for GetProposals and GetProposalById
func getProposalById(id uint64) *json.Node {
	proposal := mustGetProposal(id)
	return createProposalJsonNode(id, proposal)
}

// GetVoteStatusFromProposalById returns the vote status(max, yes, no) of a proposal.
func GetVoteStatusFromProposalById(id uint64) string {
	updateProposalsState()

	_, exists := proposals.Get(formatUint(id))
	if !exists {
		return ""
	}

	votesObj := metaNode()
	votesObj.AppendObject("proposalId", json.StringNode("proposalId", formatUint(id)))
	votesObj.AppendObject("votes", json.StringNode("votes", b64Encode(getProposalVotes(id)))) // max, yes, no

	return marshal(votesObj)
}

// GetVotesByAddress returns all votes of an address.
// included information:
// - proposalId
// - vote (yes/no)
// - weight
// - height
// - timestamp
func GetVotesByAddress(addr std.Address) string {
	updateProposalsState()

	votesObj := metaNode()
	votesArr := json.ArrayNode("", nil)

	// sertUesrVote does add data to `userVotes` with key (voter_pid) as key
	// therefore this function can not search for 'specific_address'
	userVotes.Iterate("", "", func(key string, value interface{}) bool {
		if strings.HasPrefix(key, addr.String()) {
			voteObj := createVoteJsonNode(addr, proposalId, value.(voteWithWeight))
			votesArr.AppendArray(voteObj)
		}
		return true
	})
	votesObj.AppendObject("votes", votesArr)

	return marshal(votesObj)
}

// GetVoteByAddressFromProposalById returns the vote of an address from a certain proposal.
func GetVoteByAddressFromProposalById(addr std.Address, id uint64) string {
	updateProposalsState()

	vote, exists := getUserVote(addr, id)
	if !exists {
		return ""
	}

	votesObj := metaNode()
	voteArr := json.ArrayNode("", nil)
	voteObj := createVoteJsonNode(addr, id, vote)
	voteArr.AppendArray(voteObj)
	votesObj.AppendObject("votes", voteArr)

	return marshal(votesObj)
}

func createVoteJsonNode(addr std.Address, id uint64, vote voteWithWeight) *json.Node {
	return json.Builder().
		WriteString("proposalId", formatUint(id)).
		WriteString("voteYes", formatBool(vote.Yes)).
		WriteString("voteWeight", formatUint(vote.Weight)).
		WriteString("voteHeight", formatUint(vote.VotedHeight)).
		WriteString("voteTimestamp", formatUint(vote.VotedAt)).
		Node()
}

// getProposalExtraData returns the extra data of a proposal based on its type.
func getProposalExtraData(proposalId uint64) string {
	proposal, exist := proposals.Get(formatUint(proposalId))
	if !exist {
		return ""
	}

	switch proposal.(ProposalInfo).ProposalType {
	case Text:
		return ""
	case CommunityPoolSpend:
		return getCommunityPoolSpendProposalData(proposalId)
	case ParameterChange:
		return getParameterChangeProposalData(proposalId)
	}

	return ""
}

// community pool has three extra data
// 1. to
// 2. tokenPath
// 3. amount
func getCommunityPoolSpendProposalData(proposalId uint64) string {
	proposal := mustGetProposal(proposalId)

	proposalObj := json.Builder().
		WriteString("to", proposal.CommunityPoolSpend.To.String()).
		WriteString("tokenPath", proposal.CommunityPoolSpend.TokenPath).
		WriteString("amount", formatUint(proposal.CommunityPoolSpend.Amount)).
		Node()

	return marshal(proposalObj)
}

// parameter change proposal has three extra data
func getParameterChangeProposalData(proposalId uint64) string {
	proposal := mustGetProposal(proposalId)

	msgs := proposal.Execution.Msgs
	msgsStr := strings.Join(msgs, "*GOV*")

	return msgsStr
}

// getProposalStatus returns the status of a proposal.
func getProposalStatus(id uint64) string {
	prop, exist := proposals.Get(formatUint(id))
	if !exist {
		return ""
	}
	proposal := prop.(ProposalInfo)

	config := GetConfigVersion(proposal.ConfigVersion)

	votingStart := proposal.State.CreatedAt + config.VotingStartDelay
	votingEnd := votingStart + config.VotingPeriod

	node := createProposalStateNode(proposal.State, votingStart, votingEnd)
	return marshal(node)
}

func createProposalStateNode(state ProposalState, votingStart, votingEnd uint64) *json.Node {
	return json.Builder().
		WriteString("createdAt", formatUint(state.CreatedAt)).
		WriteString("upcoming", formatBool(state.Upcoming)).
		WriteString("active", formatBool(state.Active)).
		WriteString("votingStart", formatUint(votingStart)).
		WriteString("votingEnd", formatUint(votingEnd)).
		WriteString("passed", formatBool(state.Passed)).
		WriteString("passedAt", formatUint(state.PassedAt)).
		WriteString("rejected", formatBool(state.Rejected)).
		WriteString("rejectedAt", formatUint(state.RejectedAt)).
		WriteString("canceled", formatBool(state.Canceled)).
		WriteString("canceledAt", formatUint(state.CanceledAt)).
		WriteString("executed", formatBool(state.Executed)).
		WriteString("executedAt", formatUint(state.ExecutedAt)).
		WriteString("expired", formatBool(state.Expired)).
		WriteString("expiredAt", formatUint(state.ExpiredAt)).
		Node()
}

// getProposalVotes returns the votes of a proposal.
func getProposalVotes(id uint64) string {
	prop, exist := proposals.Get(formatUint(id))
	if !exist {
		return ""
	}

	proposal := prop.(ProposalInfo)
	maxVoting := proposal.MaxVotingWeight.ToString()

	proposalObj := json.Builder().
		WriteString("quorum", formatUint(proposal.QuorumAmount)).
		WriteString("max", maxVoting).
		WriteString("yes", proposal.Yea.ToString()).
		WriteString("no", proposal.Nay.ToString()).
		Node()

	return marshal(proposalObj)
}

func metaNode() *json.Node {
	height := std.GetHeight()
	now := time.Now().Unix()

	return json.Builder().
		WriteString("height", strconv.FormatInt(height, 10)).
		WriteString("now", strconv.FormatInt(now, 10)).
		Node()
}
