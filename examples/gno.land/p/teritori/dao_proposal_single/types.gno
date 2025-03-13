package dao_proposal_single

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/json"
	dao_interfaces "gno.land/p/teritori/dao_interfaces"
	"gno.land/p/teritori/dao_utils"
	"gno.land/p/teritori/jsonutil"
)

type Ballot struct {
	Power     uint64
	Vote      Vote
	Rationale string
}

func (b Ballot) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"power":     jsonutil.Uint64Node(b.Power),
		"vote":      b.Vote.ToJSON(),
		"rationale": json.StringNode("", b.Rationale),
	})
}

func (b Ballot) String() string {
	return b.Vote.String() + " with power " + strconv.FormatUint(b.Power, 10) + "" + ": " + b.Rationale
}

type Votes struct {
	Yes     uint64
	No      uint64
	Abstain uint64
}

func (v *Votes) Add(vote Vote, power uint64) {
	switch vote {
	case VoteYes:
		v.Yes += power
	case VoteNo:
		v.No += power
	case VoteAbstain:
		v.Abstain += power
	default:
		panic("unknown vote kind")
	}
}

func (v *Votes) Remove(vote Vote, power uint64) {
	switch vote {
	case VoteYes:
		v.Yes -= power
	case VoteNo:
		v.No -= power
	case VoteAbstain:
		v.Abstain -= power
	default:
		panic("unknown vote kind")
	}
}

func (v *Votes) Total() uint64 {
	return v.Yes + v.No + v.Abstain
}

func (v Votes) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"yes":     jsonutil.Uint64Node(v.Yes),
		"no":      jsonutil.Uint64Node(v.No),
		"abstain": jsonutil.Uint64Node(v.Abstain),
	})
}

type Proposal struct {
	ID              int
	Title           string
	Description     string
	Proposer        std.Address
	StartHeight     int64
	MinVotingPeriod dao_utils.Expiration
	Expiration      dao_utils.Expiration
	Threshold       Threshold
	TotalPower      uint64
	Messages        []dao_interfaces.ExecutableMessage
	Status          ProposalStatus
	Votes           Votes
	AllowRevoting   bool

	// not in DA0-DA0 implementation:

	Ballots *avl.Tree
}

type messageWithType struct {
	Type    string
	Message dao_interfaces.ExecutableMessage
}

func (m *messageWithType) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"type":    json.StringNode("", m.Type),
		"payload": m.Message.ToJSON(),
	})
}

func formatMessages(messages []dao_interfaces.ExecutableMessage) []*json.Node {
	out := make([]*json.Node, len(messages))
	for i, m := range messages {
		mt := messageWithType{
			Type:    m.Type(),
			Message: m,
		}
		out[i] = mt.ToJSON()
	}

	return out
}

func (p Proposal) ToJSON() *json.Node {
	fields := map[string]*json.Node{
		"id":            jsonutil.IntNode(p.ID),
		"title":         json.StringNode("", p.Title),
		"description":   json.StringNode("", p.Description),
		"proposer":      jsonutil.AddressNode(p.Proposer),
		"startHeight":   jsonutil.Int64Node(p.StartHeight),
		"totalPower":    jsonutil.Uint64Node(p.TotalPower),
		"messages":      json.ArrayNode("", formatMessages(p.Messages)),
		"status":        p.Status.ToJSON(),
		"votes":         p.Votes.ToJSON(),
		"allowRevoting": json.BoolNode("", p.AllowRevoting),
		"ballots":       jsonutil.AVLTreeNode(p.Ballots, func(v interface{}) *json.Node { return v.(Ballot).ToJSON() }),
	}

	if p.MinVotingPeriod != nil {
		fields["minVotingPeriod"] = p.MinVotingPeriod.ToJSON()
	}

	if p.Expiration != nil {
		fields["expiration"] = p.Expiration.ToJSON()
	}

	if p.Threshold != nil {
		fields["threshold"] = p.Threshold.ToJSON()
	}

	return json.ObjectNode("", fields)
}

type ProposalStatus int

const (
	ProposalStatusOpen ProposalStatus = iota
	ProposalStatusPassed
	ProposalStatusExecuted
)

func (p ProposalStatus) ToJSON() *json.Node {
	return json.StringNode("", p.String())
}

func (p ProposalStatus) String() string {
	switch p {
	case ProposalStatusOpen:
		return "Open"
	case ProposalStatusPassed:
		return "Passed"
	case ProposalStatusExecuted:
		return "Executed"
	default:
		return "Unknown(" + strconv.Itoa(int(p)) + ")"
	}
}

type Vote int

const (
	VoteYes Vote = iota
	VoteNo
	VoteAbstain
)

func (v Vote) ToJSON() *json.Node {
	return json.StringNode("", v.String())
}

func (v *Vote) FromJSON(ast *json.Node) {
	v.FromString(ast.MustString())
}

func (v *Vote) FromString(s string) {
	switch s {
	case "Yes":
		*v = VoteYes
	case "No":
		*v = VoteNo
	case "Abstain":
		*v = VoteAbstain
	default:
		panic("unknown vote kind")
	}
}

func (v Vote) String() string {
	switch v {
	case VoteYes:
		return "Yes"
	case VoteNo:
		return "No"
	case VoteAbstain:
		return "Abstain"
	default:
		return "Unknown(" + strconv.Itoa(int(v)) + ")"
	}
}
