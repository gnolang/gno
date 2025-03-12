package commondao

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestProposalNew(t *testing.T) {
	id := uint64(1)
	creator := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	votingPeriod := time.Minute * 10

	p, err := NewProposal(id, creator, testPropDef{votingPeriod: votingPeriod})

	uassert.NoError(t, err)
	uassert.Equal(t, p.ID(), id)
	uassert.NotEqual(t, p.Definition(), nil)
	uassert.Equal(t, p.Quorum(), DefaultQuorum)
	uassert.True(t, p.Status() == StatusActive)
	uassert.Equal(t, p.Creator(), creator)
	uassert.False(t, p.CreatedAt().IsZero())
	uassert.NotEqual(t, p.VotingRecord(), nil)
	uassert.Empty(t, p.StatusReason())
	uassert.True(t, p.VotingDeadline() == p.CreatedAt().Add(votingPeriod))
}

func TestProposalVoteChoices(t *testing.T) {
	cases := []struct {
		name       string
		definition ProposalDefinition
		choices    []VoteChoice
	}{
		{
			name:       "custom choices",
			definition: testPropDef{voteChoices: []VoteChoice{"FOO", "BAR", "BAZ"}},
			choices:    []VoteChoice{"BAR", "BAZ", "FOO"},
		},
		{
			name:       "defaults because of empty custom choice list",
			definition: testPropDef{voteChoices: []VoteChoice{}},
			choices:    []VoteChoice{ChoiceAbstain, ChoiceNo, ChoiceYes},
		},
		{
			name:       "defaults because of single custom choice list",
			definition: testPropDef{voteChoices: []VoteChoice{"FOO"}},
			choices:    []VoteChoice{ChoiceAbstain, ChoiceNo, ChoiceYes},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := NewProposal(1, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", testPropDef{
				voteChoices: tc.choices,
			})

			choices := p.VoteChoices()

			urequire.Equal(t, len(choices), len(tc.choices), "expect vote choice count to match")
			for i, c := range choices {
				urequire.True(t, tc.choices[i] == c, "expect vote choice to match")
			}
		})
	}
}

type testPropDef struct {
	votingPeriod                time.Duration
	tallyResult                 bool
	validationErr, executionErr error
	voteChoices                 []VoteChoice
}

func (testPropDef) Title() string                          { return "" }
func (testPropDef) Body() string                           { return "" }
func (testPropDef) Quorum() float64                        { return 0 }
func (d testPropDef) VotingPeriod() time.Duration          { return d.votingPeriod }
func (d testPropDef) Validate() error                      { return d.validationErr }
func (d testPropDef) Tally(ReadOnlyVotingRecord, int) bool { return d.tallyResult }
func (d testPropDef) Execute() error                       { return d.executionErr }

func (d testPropDef) CustomVoteChoices() []VoteChoice {
	if len(d.voteChoices) > 0 {
		return d.voteChoices
	}
	return []VoteChoice{ChoiceYes, ChoiceNo, ChoiceAbstain}
}
