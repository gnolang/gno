package dao_proposal_single

import (
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/json"
	dao_interfaces "gno.land/p/teritori/dao_interfaces"
	"gno.land/p/teritori/dao_utils"
)

type NoopMessage struct{}

var _ dao_interfaces.ExecutableMessage = (*NoopMessage)(nil)

func (m NoopMessage) String() string {
	return "noop"
}

func (m NoopMessage) Type() string {
	return "noop-type"
}

func (m NoopMessage) ToJSON() *json.Node {
	return json.StringNode("", m.String())
}

func (m NoopMessage) FromJSON(ast *json.Node) {
	if ast.MustString() != m.String() {
		panic("invalid noop message")
	}
}

func TestThresholdToJSON(t *testing.T) {
	tt := PercentageThresholdPercent(1)
	tq := PercentageThresholdPercent(1)
	threshold := &ThresholdThresholdQuorum{
		Threshold: &tt,
		Quorum:    &tq,
	}
	str := threshold.ToJSON().String()
	expected := `{"thresholdQuorum":{"threshold":{"percent":1},"quorum":{"percent":1}}}`
	if expected != str {
		t.Fatalf("JSON does not match, expected %s, got %s", expected, str)
	}
}

func TestThresholdFromJSON(t *testing.T) {
	s := `   {"thresholdQuorum":    {"threshold"   :{"percent":1},"quorum":   {"percent":1}}}`
	ival := ThresholdFromJSON(json.Must(json.Unmarshal([]byte(s))))

	val, ok := ival.(*ThresholdThresholdQuorum)
	if !ok {
		t.Fatalf("expected ThresholdQuorum")
	}

	tt, ok := val.Threshold.(*PercentageThresholdPercent)
	if !ok {
		t.Fatalf("expected Percent in threshold")
	}

	if *tt != 1 {
		t.Fatalf("expected 1, got %d", *tt)
	}

	tq, ok := val.Quorum.(*PercentageThresholdPercent)
	if !ok {
		t.Fatalf("expected Percent in quorum")
	}

	if *tq != 1 {
		t.Fatalf("expected 1, got %d", *tq)
	}
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
			Expiration: dao_utils.ExpirationAtHeight(1000),
			Ballots:    avl.NewTree(),
		},
		{
			ID:          1,
			Title:       "Prop #1",
			Description: `Wolol1\"`,
			Proposer:    "0x1234567890",
			Status:      ProposalStatusExecuted,
			Expiration:  dao_utils.ExpirationAtHeight(2000),
			Messages:    []dao_interfaces.ExecutableMessage{NoopMessage{}, NoopMessage{}, NoopMessage{}},
		},
	}

	props[0].Ballots.Set("0x1234567890", Ballot{Power: 1, Vote: VoteYes, Rationale: "test"})
	iSlice := make([]*json.Node, len(props))
	for i, p := range props {
		iSlice[i] = p.ToJSON()
	}

	str := json.ArrayNode("", iSlice).String()
	expected := `[{"id":"0","title":"Prop #0","description":"Wolol0\n\t\r","proposer":"0x1234567890","startHeight":"0","totalPower":"0","messages":[],"status":"Open","votes":{"yes":"7","no":"21","abstain":"42"},"allowRevoting":false,"ballots":{"0x1234567890":{"power":"1","vote":"Yes","rationale":"test"}},"expiration":{"atHeight":"1000"}},{"id":"1","title":"Prop #1","description":"Wolol1\\\"","proposer":"0x1234567890","startHeight":"0","totalPower":"0","messages":[{"type":"noop-type","payload":"noop"},{"type":"noop-type","payload":"noop"},{"type":"noop-type","payload":"noop"}],"status":"Executed","votes":{"yes":"0","no":"0","abstain":"0"},"allowRevoting":false,"ballots":{},"expiration":{"atHeight":"2000"}}]`
	if expected != str {
		t.Fatalf("JSON does not match, expected %s, got %s", expected, str)
	}
}

func TestConfig(t *testing.T) {
	core := dao_interfaces.NewDummyCore()
	tt := PercentageThresholdPercent(1)
	tq := PercentageThresholdPercent(1)
	mod := NewDAOProposalSingle(core, &DAOProposalSingleOpts{
		MaxVotingPeriod: dao_utils.DurationHeight(42),
		MinVotingPeriod: dao_utils.DurationHeight(21),
		Threshold: &ThresholdThresholdQuorum{
			Threshold: &tt,
			Quorum:    &tq,
		},
	})
	conf := mod.ConfigJSON()
	expected := `{"threshold":{"thresholdQuorum":{"threshold":{"percent":1},"quorum":{"percent":1}}},"maxVotingPeriod":{"height":42},"minVotingPeriod":{"height":21},"onlyMembersExecute":false,"allowRevoting":false,"closeProposalOnExecutionFailure":false}`
	if expected != conf {
		t.Fatalf("Config JSON does not match, expected %s, got %s", expected, conf)
	}
}
