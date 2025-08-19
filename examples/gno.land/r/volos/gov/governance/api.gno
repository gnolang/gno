package governance

import (
	"std"
	"strconv"

	"gno.land/p/demo/json"
)

// ApiGetProposal returns a proposal as JSON object
func ApiGetProposal(proposalID string) string {
	id, err := strconv.ParseUint(proposalID, 10, 64)
	if err != nil {
		return marshalError("invalid proposal ID")
	}

	proposal := GetProposal(id)
	if proposal == nil {
		return marshalError("proposal not found")
	}

	rpcProposal := ProposalToRpc(proposal)
	return marshal(rpcProposal.JSON())
}

// ApiGetMemberSet returns the member set as JSON object
func ApiGetMemberSet() string {
	memberSet := MemberSet()
	rpcMemberSet := MemberSetToRpc(memberSet)
	return marshal(rpcMemberSet.JSON())
}

// ApiGetProposalThreshold returns the proposal threshold as JSON object
func ApiGetProposalThreshold() string {
	threshold := ProposalThreshold()
	thresholdNode := json.ObjectNode("", map[string]*json.Node{
		"proposalThreshold": json.NumberNode("proposalThreshold", float64(threshold)),
	})
	return marshal(thresholdNode)
}

// ApiGetVotingPowerQuorum returns the voting power quorum as JSON object
func ApiGetVotingPowerQuorum() string {
	quorum := VotingPowerQuorum()
	quorumNode := json.ObjectNode("", map[string]*json.Node{
		"votingPowerQuorum": json.NumberNode("votingPowerQuorum", float64(quorum)),
	})
	return marshal(quorumNode)
}

// ApiGetUserInfo returns user information as JSON object
func ApiGetUserInfo(userAddr string) string {
	addr := std.Address(userAddr)
	if !addr.IsValid() {
		return marshalError("invalid address")
	}

	userInfo := UserInfoToRpc(addr)
	return marshal(userInfo.JSON())
}

// ApiGetUserActiveVotes returns all active proposals that the user has voted on
func ApiGetUserActiveVotes(userAddr string) string {
	addr := std.Address(userAddr)
	if !addr.IsValid() {
		return marshalError("invalid address")
	}
	list := GetUserActiveProposals(addr)
	rpc := UserActiveVotesToRpc(addr, list)
	return marshal(rpc.JSON())
}

func marshal(node *json.Node) string {
	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}
	return string(b)
}

func marshalError(message string) string {
	errorNode := json.ObjectNode("", map[string]*json.Node{
		"error": json.StringNode("error", message),
	})
	return marshal(errorNode)
}
