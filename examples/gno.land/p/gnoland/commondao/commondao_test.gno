package commondao

import (
	"errors"
	"std"
	"testing"
	"time"

	"gno.land/p/demo/seqid"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestNew(t *testing.T) {
	cases := []struct {
		name    string
		parent  *CommonDAO
		members []std.Address
	}{
		{
			name:    "with parent",
			parent:  New(),
			members: []std.Address{"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},
		},
		{
			name:    "without parent",
			members: []std.Address{"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},
		},
		{
			name: "multiple members",
			members: []std.Address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
				"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
			},
		},
		{
			name: "no members",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			membersCount := len(tc.members)
			options := []Option{WithParent(tc.parent)}
			for _, m := range tc.members {
				options = append(options, WithMember(m))
			}

			dao := New(options...)

			if tc.parent == nil {
				uassert.Equal(t, nil, dao.Parent())
			} else {
				uassert.NotEqual(t, nil, dao.Parent())
			}

			urequire.Equal(t, membersCount, dao.Members().Size(), "dao members")

			var i int
			dao.Members().IterateByOffset(0, membersCount, func(addr std.Address) bool {
				uassert.Equal(t, tc.members[i], addr)
				i++
				return false
			})
		})
	}
}

func TestCommonDAOMembersAdd(t *testing.T) {
	member := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	dao := New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"))

	added := dao.Members().Add(member)
	urequire.True(t, added)

	uassert.Equal(t, 2, dao.Members().Size())
	uassert.True(t, dao.Members().Has(member))

	added = dao.Members().Add(member)
	urequire.False(t, added)
}

func TestCommonDAOMembersRemove(t *testing.T) {
	member := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	dao := New(WithMember(member))

	removed := dao.Members().Remove(member)
	urequire.True(t, removed)

	removed = dao.Members().Remove(member)
	urequire.False(t, removed)
}

func TestCommonDAOMembersHas(t *testing.T) {
	cases := []struct {
		name   string
		member std.Address
		dao    *CommonDAO
		want   bool
	}{
		{
			name:   "member",
			member: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			dao:    New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")),
			want:   true,
		},
		{
			name:   "not a dao member",
			member: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			dao:    New(WithMember("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc")),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.dao.Members().Has(tc.member)
			uassert.Equal(t, got, tc.want)
		})
	}
}

func TestCommonDAOPropose(t *testing.T) {
	cases := []struct {
		name    string
		setup   func() *CommonDAO
		creator std.Address
		def     ProposalDefinition
		err     error
	}{
		{
			name:    "ok",
			setup:   func() *CommonDAO { return New() },
			creator: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			def:     testPropDef{},
		},
		{
			name:  "nil definition",
			setup: func() *CommonDAO { return New() },
			err:   ErrProposalDefinitionRequired,
		},
		{
			name:  "invalid creator address",
			setup: func() *CommonDAO { return New() },
			def:   testPropDef{},
			err:   ErrInvalidCreatorAddress,
		},
		{
			name: "proposal ID overflow",
			setup: func() *CommonDAO {
				dao := New()
				dao.genID = seqid.ID(1<<64 - 1)
				return dao
			},
			creator: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			def:     testPropDef{},
			err:     ErrOverflow,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dao := tc.setup()

			p, err := dao.Propose(tc.creator, tc.def)

			if tc.err != nil {
				urequire.ErrorIs(t, err, tc.err)
				return
			}

			urequire.NoError(t, err)

			_, found := dao.GetActiveProposal(p.ID())
			urequire.True(t, found, "proposal not found")
			uassert.Equal(t, p.Creator(), tc.creator)
		})
	}
}

func TestCommonDAOVote(t *testing.T) {
	cases := []struct {
		name       string
		setup      func() *CommonDAO
		member     std.Address
		choice     VoteChoice
		proposalID uint64
		err        error
	}{
		{
			name: "ok",
			setup: func() *CommonDAO {
				member := std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := New(WithMember(member))
				dao.Propose(member, testPropDef{})
				return dao
			},
			member:     "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			choice:     ChoiceYes,
			proposalID: 1,
		},
		{
			name:   "invalid vote choice",
			setup:  func() *CommonDAO { return New() },
			choice: VoteChoice("invalid"),
			err:    ErrInvalidVoteChoice,
		},
		{
			name:   "not a member",
			setup:  func() *CommonDAO { return New() },
			member: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			err:    ErrNotMember,
		},
		{
			name: "proposal not found",
			setup: func() *CommonDAO {
				return New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"))
			},
			member:     "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			proposalID: 42,
			err:        ErrProposalNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dao := tc.setup()

			err := dao.Vote(tc.member, tc.proposalID, tc.choice)

			if tc.err != nil {
				urequire.ErrorIs(t, err, tc.err)
				return
			}

			urequire.NoError(t, err)

			p, found := dao.GetActiveProposal(tc.proposalID)
			urequire.True(t, found, "proposal not found")

			record := p.VotingRecord()
			uassert.True(t, record.HasVoted(tc.member))
			uassert.Equal(t, record.VoteCount(tc.choice), 1)
		})
	}
}

func TestCommonDAOTally(t *testing.T) {
	cases := []struct {
		name  string
		dao   *CommonDAO
		votes []Vote
		err   error
	}{
		{
			name: "pass",
			dao: New(
				WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
				WithMember("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc"),
				WithMember("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			),
			votes: []Vote{
				{Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", Choice: ChoiceYes},
				{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes},
			},
		},
		{
			name: "no votes",
			dao:  New(),
			err:  ErrLowParticipation,
		},
		{
			name: "no quorum",
			dao: New(
				WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
				WithMember("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc"),
				WithMember("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			),
			votes: []Vote{
				{Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", Choice: ChoiceYes},
			},
			err: ErrLowParticipation,
		},
		{
			name: "no consensus",
			dao: New(
				WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
				WithMember("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc"),
				WithMember("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			),
			votes: []Vote{
				{Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", Choice: ChoiceYes},
				{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceNo},
			},
			err: ErrNoConcensus,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := tc.dao.Propose("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", majorityPropDef{})
			for _, v := range tc.votes {
				tc.dao.Vote(v.Address, p.ID(), v.Choice)
			}

			err := tc.dao.Tally(p.ID())

			if tc.err != nil {
				uassert.Error(t, err, "expect an error")
			} else {
				uassert.NoError(t, err, "expect no error")
			}
		})
	}
}

func TestCommonDAOExecute(t *testing.T) {
	errValidation := errors.New("validation error")
	errExecution := errors.New("execution error")
	cases := []struct {
		name         string
		setup        func() *CommonDAO
		proposalID   uint64
		status       ProposalStatus
		statusReason string
		err          error
	}{
		{
			name: "ok",
			setup: func() *CommonDAO {
				members := []std.Address{
					"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
					"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
					"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				}
				dao := New(WithMember(members[0]), WithMember(members[1]), WithMember(members[2]))
				p, _ := dao.Propose(members[0], testPropDef{tallyResult: true})
				p.record.AddVote(members[0], ChoiceYes)
				p.record.AddVote(members[1], ChoiceYes)
				return dao
			},
			status:     StatusExecuted,
			proposalID: 1,
		},
		{
			name:       "proposal not found",
			setup:      func() *CommonDAO { return New() },
			proposalID: 1,
			err:        ErrProposalNotFound,
		},
		{
			name: "proposal not active",
			setup: func() *CommonDAO {
				member := std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := New(WithMember(member))
				p, _ := dao.Propose(member, testPropDef{})
				p.status = StatusExecuted
				return dao
			},
			proposalID: 1,
			err:        ErrStatusIsNotActive,
		},
		{
			name: "voting deadline not met",
			setup: func() *CommonDAO {
				member := std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := New(WithMember(member))
				dao.Propose(member, testPropDef{votingPeriod: time.Minute * 5})
				return dao
			},
			proposalID: 1,
			err:        ErrVotingDeadlineNotMet,
		},
		{
			name: "validation error",
			setup: func() *CommonDAO {
				member := std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := New(WithMember(member))
				dao.Propose(member, testPropDef{validationErr: errValidation})
				return dao
			},
			proposalID:   1,
			status:       StatusFailed,
			statusReason: errValidation.Error(),
		},
		{
			name: "tally error",
			setup: func() *CommonDAO {
				member := std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := New(WithMember(member))
				dao.Propose(member, testPropDef{})
				return dao
			},
			proposalID:   1,
			status:       StatusFailed,
			statusReason: ErrLowParticipation.Error(),
		},
		{
			name: "execution error",
			setup: func() *CommonDAO {
				members := []std.Address{
					"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
					"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
					"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				}
				dao := New(WithMember(members[0]), WithMember(members[1]), WithMember(members[2]))
				p, _ := dao.Propose(members[0], testPropDef{
					tallyResult:  true,
					executionErr: errExecution,
				})
				p.record.AddVote(members[0], ChoiceYes)
				p.record.AddVote(members[1], ChoiceYes)
				return dao
			},
			proposalID:   1,
			status:       StatusFailed,
			statusReason: errExecution.Error(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dao := tc.setup()

			err := dao.Execute(tc.proposalID)

			if tc.err != nil {
				urequire.ErrorIs(t, err, tc.err, "expect error to match")
				return
			}

			urequire.NoError(t, err, "expect no error")

			_, found := dao.GetActiveProposal(tc.proposalID)
			urequire.False(t, found, "proposal should not be active")

			p, found := dao.GetFinishedProposal(tc.proposalID)
			urequire.True(t, found, "proposal must be found")
			uassert.Equal(t, string(p.Status()), string(tc.status), "status must match")
			uassert.Equal(t, string(p.StatusReason()), string(tc.statusReason), "status reason must match")
		})
	}
}

type majorityPropDef struct{ testPropDef }

func (majorityPropDef) Tally(r ReadOnlyVotingRecord, membersCount int) bool {
	_, success := SelectChoiceByAbsoluteMajority(r, membersCount)
	return success
}
