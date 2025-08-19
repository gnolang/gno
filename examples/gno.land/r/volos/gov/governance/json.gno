package governance

import (
	"std"

	"gno.land/p/demo/json"
	"gno.land/p/nt/commondao"
	"gno.land/r/volos/gov/vls"
	"gno.land/r/volos/gov/xvls"
)

// RpcProposal represents a proposal in RPC format
type RpcProposal struct {
	ID               uint64 `json:"id"`
	Title            string `json:"title"`
	Body             string `json:"body"`
	Status           string `json:"status"`
	VotingDeadline   int64  `json:"votingDeadline"`
	Proposer         string `json:"proposer"`
	QuorumAtCreation int64  `json:"quorumAtCreation"`
}

func ProposalToRpc(proposal *commondao.Proposal) RpcProposal {
	if proposal == nil {
		return RpcProposal{}
	}

	def := proposal.Definition().(VolosProposalDefinition)
	return RpcProposal{
		ID:               proposal.ID(),
		Title:            def.Title(),
		Body:             def.Body(),
		Status:           string(proposal.Status()),
		VotingDeadline:   proposal.VotingDeadline().Unix(),
		Proposer:         proposal.Creator().String(),
		QuorumAtCreation: def.QuorumAtCreation,
	}
}

func (r RpcProposal) JSON() *json.Node {
	return json.ObjectNode("proposal", map[string]*json.Node{
		"id":               json.NumberNode("id", float64(r.ID)),
		"title":            json.StringNode("title", r.Title),
		"body":             json.StringNode("body", r.Body),
		"status":           json.StringNode("status", r.Status),
		"votingDeadline":   json.NumberNode("votingDeadline", float64(r.VotingDeadline)),
		"proposer":         json.StringNode("proposer", r.Proposer),
		"quorumAtCreation": json.NumberNode("quorumAtCreation", float64(r.QuorumAtCreation)),
	})
}

// RpcMemberSet represents the member set in RPC format
type RpcMemberSet struct {
	Members []string `json:"members"`
	Size    int      `json:"size"`
}

func MemberSetToRpc(memberSet commondao.MemberSet) RpcMemberSet {
	members := []string{}
	memberSet.IterateByOffset(0, memberSet.Size(), func(addr std.Address) bool {
		members = append(members, addr.String())
		return false
	})

	return RpcMemberSet{
		Members: members,
		Size:    memberSet.Size(),
	}
}

func (r RpcMemberSet) JSON() *json.Node {
	memberNodes := []*json.Node{}
	for _, member := range r.Members {
		memberNodes = append(memberNodes, json.StringNode("", member))
	}

	return json.ObjectNode("memberSet", map[string]*json.Node{
		"members": json.ArrayNode("members", memberNodes),
		"size":    json.NumberNode("size", float64(r.Size)),
	})
}

// RpcUserInfo represents user information in RPC format
type RpcUserInfo struct {
	Address           string `json:"address"`
	VlsBalance        int64  `json:"vlsBalance"`
	XvlsBalance       int64  `json:"xvlsBalance"`
	ProposalThreshold int64  `json:"proposalThreshold"`
	IsMember          bool   `json:"isMember"`
}

func UserInfoToRpc(userAddr std.Address) RpcUserInfo {
	return RpcUserInfo{
		Address:           userAddr.String(),
		VlsBalance:        vls.BalanceOf(userAddr),
		XvlsBalance:       xvls.BalanceOf(userAddr),
		ProposalThreshold: ProposalThreshold(),
		IsMember:          volosGovernance.Members().Has(userAddr),
	}
}

func (r RpcUserInfo) JSON() *json.Node {
	return json.ObjectNode("userInfo", map[string]*json.Node{
		"address":           json.StringNode("address", r.Address),
		"vlsBalance":        json.NumberNode("vlsBalance", float64(r.VlsBalance)),
		"xvlsBalance":       json.NumberNode("xvlsBalance", float64(r.XvlsBalance)),
		"proposalThreshold": json.NumberNode("proposalThreshold", float64(r.ProposalThreshold)),
		"isMember":          json.BoolNode("isMember", r.IsMember),
	})
}

// RpcUserActiveVotes lists active proposals a user voted on
type RpcUserActiveVotes struct {
	User      string        `json:"user"`
	Proposals []RpcProposal `json:"proposals"`
}

func UserActiveVotesToRpc(user std.Address, proposals []*commondao.Proposal) RpcUserActiveVotes {
	items := make([]RpcProposal, 0, len(proposals))
	for _, p := range proposals {
		items = append(items, ProposalToRpc(p))
	}
	return RpcUserActiveVotes{
		User:      user.String(),
		Proposals: items,
	}
}

func (r RpcUserActiveVotes) JSON() *json.Node {
	arr := []*json.Node{}
	for _, p := range r.Proposals {
		arr = append(arr, p.JSON())
	}
	return json.ObjectNode("userActiveVotes", map[string]*json.Node{
		"user":      json.StringNode("user", r.User),
		"proposals": json.ArrayNode("proposals", arr),
	})
}
