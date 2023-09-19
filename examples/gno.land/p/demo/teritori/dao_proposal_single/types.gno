package dao_proposal_single

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	dao_interfaces "gno.land/p/demo/teritori/dao_interfaces"
	"gno.land/p/demo/teritori/ujson"
)

type Ballot struct {
	Power     uint64
	Vote      Vote
	Rationale string
}

func (b Ballot) ToJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "power", Value: b.Power},
		{Key: "vote", Value: b.Vote},
		{Key: "rationale", Value: b.Rationale},
	})
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

func (v Votes) ToJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "yes", Value: v.Yes},
		{Key: "no", Value: v.No},
		{Key: "abstain", Value: v.Abstain},
	})
}

type Proposal struct {
	ID          int
	Title       string
	Description string
	Proposer    std.Address
	StartHeight int64
	// TODO: min_voting_period
	// TODO: expiration
	Threshold     Threshold
	TotalPower    uint64
	Messages      []dao_interfaces.ExecutableMessage
	Status        ProposalStatus
	Votes         Votes
	AllowRevoting bool

	// not in DA0-DA0 implementation:

	Ballots *avl.Tree
}

var _ ujson.JSONAble = (*Proposal)(nil)

type messageWithType struct {
	Type    string
	Message dao_interfaces.ExecutableMessage
}

func (m *messageWithType) ToJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "type", Value: m.Type},
		{Key: "payload", Value: m.Message},
	})
}

func formatMessages(messages []dao_interfaces.ExecutableMessage) string {
	var out []*messageWithType
	for _, m := range messages {
		out = append(out, &messageWithType{
			Type:    m.Type(),
			Message: m,
		})
	}
	return ujson.FormatSlice(out)
}

func (p Proposal) ToJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "id", Value: p.ID},
		{Key: "title", Value: p.Title},
		{Key: "description", Value: p.Description},
		{Key: "proposer", Value: p.Proposer},
		{Key: "startHeight", Value: p.StartHeight},
		{Key: "threshold", Value: p.Threshold},
		{Key: "totalPower", Value: p.TotalPower},
		{Key: "messages", Value: formatMessages(p.Messages), Raw: true},
		{Key: "status", Value: p.Status},
		{Key: "votes", Value: p.Votes},
		{Key: "allowRevoting", Value: p.AllowRevoting},

		{Key: "ballots", Value: p.Ballots},
	})
}

type ProposalStatus int

const (
	ProposalStatusOpen ProposalStatus = iota
	ProposalStatusPassed
	ProposalStatusExecuted
)

func (p ProposalStatus) ToJSON() string {
	return ujson.FormatString(p.String())
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

func (v Vote) ToJSON() string {
	return ujson.FormatString(v.String())
}

func (v *Vote) FromJSON(ast *ujson.JSONASTNode) {
	var val int
	ast.ParseAny(&val)
	// FIXME: validate
	*v = Vote(val)
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
