package dao_interfaces

import (
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/jsonutil_v3"
)

type NoopMessage struct{}

var _ ExecutableMessage = (*NoopMessage)(nil)

func (m NoopMessage) String() string {
	return "noop"
}

func (m NoopMessage) Binary() []byte {
	return nil
}

func (m NoopMessage) Type() string {
	return "noop-type"
}

func (m NoopMessage) ToJSON() string {
	return jsonutil.FormatString(m.String())
}

func TestProposalJSON(t *testing.T) {
	props := []Proposal{
		{
			ID:          0,
			Title:       "Prop #0",
			Description: "Wolol0\n\t\r",
			Proposer:    "0x1234567890",
			Votes: Votes{
				Yes:     7,
				No:      21,
				Abstain: 42,
			},
			Ballots: avl.NewTree(),
		},
		{
			ID:          1,
			Title:       "Prop #1",
			Description: `Wolol1\"`,
			Proposer:    "0x1234567890",
			Status:      ProposalStatusExecuted,
			Messages:    []ExecutableMessage{NoopMessage{}, NoopMessage{}, NoopMessage{}},
		},
	}
	props[0].Ballots.Set("0x1234567890", Ballot{Power: 1, Vote: VoteYes, Rationale: "test"})
	str := jsonutil.FormatSlice(props)
	expected := `[{"id":0,"title":"Prop #0","description":"Wolol0\n\t\r","proposer":"0x1234567890","messages":[],"ballots":{"0x1234567890":{"power":1,"vote":"Yes","rationale":"test"}},"votes":{"yes":7,"no":21,"abstain":42},"status":"Open"},{"id":1,"title":"Prop #1","description":"Wolol1\\\"","proposer":"0x1234567890","messages":["noop","noop","noop"],"ballots":{},"votes":{"yes":0,"no":0,"abstain":0},"status":"Executed"}]`
	if expected != str {
		t.Fatalf("JSON does not match, expected %s, got %s", expected, str)
	}
}
