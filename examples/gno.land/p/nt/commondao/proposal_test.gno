package commondao

import (
	"errors"
	"testing"
	"time"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestProposalNew(t *testing.T) {
	cases := []struct {
		name       string
		creator    address
		definition ProposalDefinition
		err        error
	}{
		{
			name:       "success",
			creator:    "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			definition: testPropDef{votingPeriod: time.Minute * 10},
		},
		{
			name:       "invalid creator address",
			creator:    "invalid",
			definition: testPropDef{},
			err:        ErrInvalidCreatorAddress,
		},
		{
			name:    "max custom vote choices exceeded",
			creator: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			definition: testPropDef{
				voteChoices: make([]VoteChoice, MaxCustomVoteChoices+1),
			},
			err: ErrMaxCustomVoteChoices,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := uint64(1)

			p, err := NewProposal(id, tc.creator, tc.definition)

			if tc.err != nil {
				urequire.ErrorIs(t, err, tc.err, "expected an error")
				return
			}

			urequire.NoError(t, err, "unexpected error")
			uassert.Equal(t, p.ID(), id)
			uassert.NotEqual(t, p.Definition(), nil)
			uassert.True(t, p.Status() == StatusActive)
			uassert.Equal(t, p.Creator(), tc.creator)
			uassert.False(t, p.CreatedAt().IsZero())
			uassert.NotEqual(t, p.VotingRecord(), nil)
			uassert.Empty(t, p.StatusReason())
			uassert.True(t, p.VotingDeadline() == p.CreatedAt().Add(tc.definition.VotingPeriod()))
		})
	}
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

func TestIsQuorumReached(t *testing.T) {
	cases := []struct {
		name    string
		quorum  float64
		members []address
		votes   []Vote
		fail    bool
	}{
		{
			name:   "one third",
			quorum: QuorumOneThird,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
			},
			votes: []Vote{
				{Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", Choice: ChoiceYes},
			},
		},
		{
			name:   "one third no quorum",
			quorum: QuorumOneThird,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
			},
			fail: true,
		},
		{
			name:   "half",
			quorum: QuorumHalf,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
				"g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt",
			},
			votes: []Vote{
				{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes},
				{Address: "g1w4ek2u3jta047h6lta047h6lta047h6l9huexc", Choice: ChoiceNo},
			},
		},
		{
			name:   "half no quorum",
			quorum: QuorumHalf,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
				"g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt",
			},
			votes: []Vote{
				{Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", Choice: ChoiceYes},
			},
			fail: true,
		},
		{
			name:   "two thirds",
			quorum: QuorumTwoThirds,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
			},
			votes: []Vote{
				{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes},
				{Address: "g1w4ek2u3jta047h6lta047h6lta047h6l9huexc", Choice: ChoiceNo},
			},
		},
		{
			name:   "two thirds no quorum",
			quorum: QuorumTwoThirds,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
			},
			votes: []Vote{
				{Address: "g1w4ek2u3jta047h6lta047h6lta047h6l9huexc", Choice: ChoiceNo},
			},
			fail: true,
		},
		{
			name:   "three fourths",
			quorum: QuorumThreeFourths,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
				"g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt",
			},
			votes: []Vote{
				{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes},
				{Address: "g1w4ek2u3jta047h6lta047h6lta047h6l9huexc", Choice: ChoiceNo},
				{Address: "g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt", Choice: ChoiceNo},
			},
		},
		{
			name:   "three fourths no quorum",
			quorum: QuorumThreeFourths,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
				"g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt",
			},
			votes: []Vote{
				{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes},
			},
			fail: true,
		},
		{
			name:   "full",
			quorum: QuorumFull,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			},
			votes: []Vote{
				{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceNo},
				{Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", Choice: ChoiceNo},
			},
		},
		{
			name:   "full no quorum",
			quorum: QuorumFull,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			},
			votes: []Vote{
				{Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", Choice: ChoiceNo},
			},
			fail: true,
		},
		{
			name:   "no quorum with empty vote",
			quorum: QuorumHalf,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			},
			votes: []Vote{
				{Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", Choice: ChoiceNone},
			},
			fail: true,
		},
		{
			name:   "no quorum with abstention",
			quorum: QuorumHalf,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			},
			votes: []Vote{
				{Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", Choice: ChoiceAbstain},
			},
			fail: true,
		},
		{
			name:   "invalid quorum percentage",
			quorum: -1,
			fail:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			members := NewMemberStorage()
			for _, m := range tc.members {
				members.Add(m)
			}

			var record VotingRecord
			for _, v := range tc.votes {
				record.AddVote(v)
			}

			success := IsQuorumReached(tc.quorum, record.Readonly(), NewMemberSet(members))

			if tc.fail {
				uassert.False(t, success, "expect quorum to fail")
			} else {
				uassert.True(t, success, "expect quorum to succeed")
			}
		})
	}
}

func TestMustValidate(t *testing.T) {
	uassert.NotPanics(t, func() {
		MustValidate(testPropDef{})
	}, "expect validation to succeed")

	uassert.PanicsWithMessage(t, "validable proposal definition is nil", func() {
		MustValidate(nil)
	}, "expect validation to panic with nil definition")

	uassert.PanicsWithMessage(t, "boom!", func() {
		MustValidate(testPropDef{validationErr: errors.New("boom!")})
	}, "expect validation to panic")
}

type testPropDef struct {
	votingPeriod            time.Duration
	tallyResult             bool
	validationErr, tallyErr error
	voteChoices             []VoteChoice
}

func (testPropDef) Title() string                 { return "" }
func (testPropDef) Body() string                  { return "" }
func (d testPropDef) VotingPeriod() time.Duration { return d.votingPeriod }
func (d testPropDef) Validate() error             { return d.validationErr }

func (d testPropDef) Tally(ReadonlyVotingRecord, MemberSet) (bool, error) {
	return d.tallyResult, d.tallyErr
}

func (d testPropDef) CustomVoteChoices() []VoteChoice {
	if len(d.voteChoices) > 0 {
		return d.voteChoices
	}
	return []VoteChoice{ChoiceYes, ChoiceNo, ChoiceAbstain}
}
