package commondao_test

import (
	"errors"
	"testing"
	"time"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"

	"gno.land/p/nt/commondao"
)

func TestProposalNew(t *testing.T) {
	cases := []struct {
		name       string
		creator    address
		definition commondao.ProposalDefinition
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
			err:        commondao.ErrInvalidCreatorAddress,
		},
		{
			name:    "max custom vote choices exceeded",
			creator: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			definition: testPropDef{
				voteChoices: make([]commondao.VoteChoice, commondao.MaxCustomVoteChoices+1),
			},
			err: commondao.ErrMaxCustomVoteChoices,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := uint64(1)

			p, err := commondao.NewProposal(id, tc.creator, tc.definition)

			if tc.err != nil {
				urequire.ErrorIs(t, err, tc.err, "expected an error")
				return
			}

			urequire.NoError(t, err, "unexpected error")
			uassert.Equal(t, p.ID(), id)
			uassert.NotEqual(t, p.Definition(), nil)
			uassert.True(t, p.Status() == commondao.StatusActive)
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
		definition commondao.ProposalDefinition
		choices    []commondao.VoteChoice
	}{
		{
			name:       "custom choices",
			definition: testPropDef{voteChoices: []commondao.VoteChoice{"FOO", "BAR", "BAZ"}},
			choices: []commondao.VoteChoice{
				"BAR",
				"BAZ",
				"FOO",
			},
		},
		{
			name:       "defaults because of empty custom choice list",
			definition: testPropDef{voteChoices: []commondao.VoteChoice{}},
			choices: []commondao.VoteChoice{
				commondao.ChoiceAbstain,
				commondao.ChoiceNo,
				commondao.ChoiceYes,
			},
		},
		{
			name:       "defaults because of single custom choice list",
			definition: testPropDef{voteChoices: []commondao.VoteChoice{"FOO"}},
			choices: []commondao.VoteChoice{
				commondao.ChoiceAbstain,
				commondao.ChoiceNo,
				commondao.ChoiceYes,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := commondao.NewProposal(1, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", testPropDef{
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
		votes   []commondao.Vote
		fail    bool
	}{
		{
			name:   "one third",
			quorum: commondao.QuorumOneThird,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
			},
			votes: []commondao.Vote{
				{
					Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
					Choice:  commondao.ChoiceYes,
				},
			},
		},
		{
			name:   "one third no quorum",
			quorum: commondao.QuorumOneThird,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
			},
			fail: true,
		},
		{
			name:   "half",
			quorum: commondao.QuorumHalf,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
				"g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt",
			},
			votes: []commondao.Vote{
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				},
				{
					Address: "g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
					Choice:  commondao.ChoiceNo,
				},
			},
		},
		{
			name:   "half no quorum",
			quorum: commondao.QuorumHalf,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
				"g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt",
			},
			votes: []commondao.Vote{
				{
					Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
					Choice:  commondao.ChoiceYes,
				},
			},
			fail: true,
		},
		{
			name:   "two thirds",
			quorum: commondao.QuorumTwoThirds,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
			},
			votes: []commondao.Vote{
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				},
				{
					Address: "g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
					Choice:  commondao.ChoiceNo,
				},
			},
		},
		{
			name:   "two thirds no quorum",
			quorum: commondao.QuorumTwoThirds,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
			},
			votes: []commondao.Vote{
				{
					Address: "g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
					Choice:  commondao.ChoiceNo,
				},
			},
			fail: true,
		},
		{
			name:   "three fourths",
			quorum: commondao.QuorumThreeFourths,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
				"g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt",
			},
			votes: []commondao.Vote{
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				},
				{
					Address: "g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
					Choice:  commondao.ChoiceNo,
				},
				{
					Address: "g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt",
					Choice:  commondao.ChoiceNo,
				},
			},
		},
		{
			name:   "three fourths no quorum",
			quorum: commondao.QuorumThreeFourths,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
				"g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt",
			},
			votes: []commondao.Vote{
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				},
			},
			fail: true,
		},
		{
			name:   "full",
			quorum: commondao.QuorumFull,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			},
			votes: []commondao.Vote{
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceNo,
				},
				{
					Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
					Choice:  commondao.ChoiceNo,
				},
			},
		},
		{
			name:   "full no quorum",
			quorum: commondao.QuorumFull,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			},
			votes: []commondao.Vote{
				{
					Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
					Choice:  commondao.ChoiceNo,
				},
			},
			fail: true,
		},
		{
			name:   "no quorum with empty vote",
			quorum: commondao.QuorumHalf,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			},
			votes: []commondao.Vote{
				{
					Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
					Choice:  commondao.ChoiceNone,
				},
			},
			fail: true,
		},
		{
			name:   "no quorum with abstention",
			quorum: commondao.QuorumHalf,
			members: []address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			},
			votes: []commondao.Vote{
				{
					Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
					Choice:  commondao.ChoiceAbstain,
				},
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
			members := commondao.NewMemberStorage()
			storage := commondao.MustNewReadonlyMemberStorage(members)
			for _, m := range tc.members {
				members.Add(m)
			}

			var record commondao.VotingRecord
			for _, v := range tc.votes {
				record.AddVote(v)
			}

			success := commondao.IsQuorumReached(tc.quorum, record.Readonly(), *storage)

			if tc.fail {
				uassert.False(t, success, "expect quorum to fail")
			} else {
				uassert.True(t, success, "expect quorum to succeed")
			}
		})
	}
}

func TestProposalTally(t *testing.T) {
	errTest := errors.New("test")
	creator := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	cases := []struct {
		name   string
		setup  func() *commondao.Proposal
		status commondao.ProposalStatus
		err    error
	}{
		{
			name: "passed",
			setup: func() *commondao.Proposal {
				p, _ := commondao.NewProposal(1, creator, testPropDef{tallyResult: true})
				return p
			},
			status: commondao.StatusPassed,
		},
		{
			name: "rejected",
			setup: func() *commondao.Proposal {
				p, _ := commondao.NewProposal(1, creator, testPropDef{tallyResult: false})
				return p
			},
			status: commondao.StatusRejected,
		},
		{
			name: "proposal is not active",
			setup: func() *commondao.Proposal {
				p, _ := commondao.NewProposal(1, creator, testPropDef{tallyResult: true})
				p.Tally(commondao.NewMemberStorage())
				return p
			},
			err: commondao.ErrStatusIsNotActive,
		},
		{
			name: "tally error",
			setup: func() *commondao.Proposal {
				p, _ := commondao.NewProposal(1, creator, testPropDef{tallyErr: errTest})
				return p
			},
			err: errTest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := tc.setup()
			members := commondao.NewMemberStorage()

			err := p.Tally(members)

			if tc.err != nil {
				urequire.ErrorIs(t, err, tc.err)
				return
			}

			urequire.NoError(t, err)
			urequire.Equal(t, string(tc.status), string(p.Status()))
		})
	}
}

func TestMustValidate(t *testing.T) {
	uassert.NotPanics(t, func() {
		commondao.MustValidate(testPropDef{})
	}, "expect validation to succeed")

	uassert.PanicsWithMessage(t, "validable proposal definition is nil", func() {
		commondao.MustValidate(nil)
	}, "expect validation to panic with nil definition")

	uassert.PanicsWithMessage(t, "boom!", func() {
		commondao.MustValidate(testPropDef{validationErr: errors.New("boom!")})
	}, "expect validation to panic")
}

// Executable non crossing proposal definition for unit tests
type testPropDef struct {
	votingPeriod            time.Duration
	tallyResult             bool
	validationErr, tallyErr error
	voteChoices             []commondao.VoteChoice
}

func (testPropDef) Title() string                 { return "" }
func (testPropDef) Body() string                  { return "" }
func (d testPropDef) VotingPeriod() time.Duration { return d.votingPeriod }
func (d testPropDef) Validate() error             { return d.validationErr }

func (d testPropDef) Tally(commondao.VotingContext) (bool, error) {
	return d.tallyResult, d.tallyErr
}

func (d testPropDef) CustomVoteChoices() []commondao.VoteChoice {
	if len(d.voteChoices) > 0 {
		return d.voteChoices
	}
	return []commondao.VoteChoice{commondao.ChoiceYes, commondao.ChoiceNo, commondao.ChoiceAbstain}
}
