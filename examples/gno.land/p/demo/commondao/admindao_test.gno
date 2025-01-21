package admindao

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
		parent  *AdminDAO
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

			urequire.Equal(t, len(tc.members), len(dao.Members()), "dao members")
			for i, m := range dao.Members() {
				uassert.Equal(t, tc.members[i], m)
			}
		})
	}
}

func TestAdminDAOAddMember(t *testing.T) {
	member := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	dao := New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"))

	err := dao.AddMember(member)
	urequire.NoError(t, err)
	uassert.Equal(t, 2, len(dao.Members()))
	uassert.True(t, dao.IsMember(member))

	err = dao.AddMember(member)
	uassert.ErrorIs(t, err, ErrMemberExists)
}

func TestAdminDAORemoveMember(t *testing.T) {
	member := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	dao := New(WithMember(member))

	removed := dao.RemoveMember(member)
	urequire.True(t, removed)

	removed = dao.RemoveMember(member)
	urequire.False(t, removed)
}

func TestAdminDAOIsMember(t *testing.T) {
	cases := []struct {
		name   string
		member std.Address
		dao    *AdminDAO
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
			got := tc.dao.IsMember(tc.member)
			uassert.Equal(t, got, tc.want)
		})
	}
}

func TestAdminDAOPropose(t *testing.T) {
	cases := []struct {
		name    string
		setup   func() *AdminDAO
		creator std.Address
		def     ProposalDefinition
		err     error
	}{
		{
			name:    "ok",
			setup:   func() *AdminDAO { return New() },
			creator: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			def:     &testPropDef{},
		},
		{
			name:  "nil definition",
			setup: func() *AdminDAO { return New() },
			err:   ErrProposalDefinitionRequired,
		},
		{
			name:  "invalid address",
			setup: func() *AdminDAO { return New() },
			def:   &testPropDef{},
			err:   ErrInvalidAddress,
		},
		{
			name: "proposal ID overflow",
			setup: func() *AdminDAO {
				dao := New()
				dao.genID = seqid.ID(1<<64 - 1)
				return dao
			},
			creator: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			def:     &testPropDef{},
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

func TestAdminDAOVote(t *testing.T) {
	cases := []struct {
		name       string
		setup      func() *AdminDAO
		member     std.Address
		choice     VoteChoice
		proposalID uint64
		err        error
	}{
		{
			name: "ok",
			setup: func() *AdminDAO {
				member := std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := New(WithMember(member))
				dao.Propose(member, &testPropDef{})
				return dao
			},
			member:     "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			choice:     ChoiceYes,
			proposalID: 1,
		},
		{
			name:   "invalid vote choice",
			setup:  func() *AdminDAO { return New() },
			choice: VoteChoice("invalid"),
			err:    ErrInvalidVoteChoice,
		},
		{
			name:   "not a member",
			setup:  func() *AdminDAO { return New() },
			member: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			err:    ErrNotMember,
		},
		{
			name: "proposal not found",
			setup: func() *AdminDAO {
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

func TestAdminDAOTally(t *testing.T) {
	cases := []struct {
		name         string
		dao          *AdminDAO
		votes        []Vote
		status       ProposalStatus
		statusReason string
		stats        Stats
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
			status: StatusActive,
			stats:  Stats{YayVotes: 2, Abstained: 1},
		},
		{
			name:         "no votes",
			dao:          New(),
			status:       StatusFailed,
			statusReason: "low participation",
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
			status:       StatusFailed,
			statusReason: "low participation",
			stats:        Stats{YayVotes: 1, Abstained: 2},
		},
		{
			name: "majority not met",
			dao: New(
				WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"),
				WithMember("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc"),
				WithMember("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			),
			votes: []Vote{
				{Address: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", Choice: ChoiceYes},
				{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceNo},
			},
			status:       StatusFailed,
			statusReason: "majority not met",
			stats:        Stats{YayVotes: 1, NayVotes: 1, Abstained: 1},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewProposal(1, "", testPropDef{})
			for _, v := range tc.votes {
				p.record.AddVote(v.Address, v.Choice)
			}

			stats := tc.dao.Tally(p)

			uassert.Equal(t, string(p.Status()), string(tc.status))
			uassert.Equal(t, p.StatusReason(), tc.statusReason)
			uassert.Equal(t, stats.YayVotes, tc.stats.YayVotes)
			uassert.Equal(t, stats.NayVotes, tc.stats.NayVotes)
			uassert.Equal(t, stats.Abstained, tc.stats.Abstained)
		})
	}
}

func TestAdminDAOExecute(t *testing.T) {
	errValidation := errors.New("validation error")
	errExecution := errors.New("execution error")
	cases := []struct {
		name         string
		setup        func() *AdminDAO
		proposalID   uint64
		status       ProposalStatus
		statusReason string
		err          error
	}{
		{
			name: "ok",
			setup: func() *AdminDAO {
				members := []std.Address{
					"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
					"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
					"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				}
				dao := New(WithMember(members[0]), WithMember(members[1]), WithMember(members[2]))
				p, _ := dao.Propose(members[0], &testPropDef{})
				p.record.AddVote(members[0], ChoiceYes)
				p.record.AddVote(members[1], ChoiceYes)
				return dao
			},
			status:     StatusExecuted,
			proposalID: 1,
		},
		{
			name:       "proposal not found",
			setup:      func() *AdminDAO { return New() },
			proposalID: 1,
			err:        ErrProposalNotFound,
		},
		{
			name: "proposal not active",
			setup: func() *AdminDAO {
				member := std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := New(WithMember(member))
				p, _ := dao.Propose(member, &testPropDef{})
				p.status = StatusExecuted
				return dao
			},
			proposalID: 1,
			err:        ErrStatusIsNotActive,
		},
		{
			name: "voting deadline not met",
			setup: func() *AdminDAO {
				member := std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := New(WithMember(member))
				dao.Propose(member, &testPropDef{votingPeriod: time.Minute * 5})
				return dao
			},
			proposalID: 1,
			err:        ErrVotingDeadlineNotMet,
		},
		{
			name: "validation error",
			setup: func() *AdminDAO {
				member := std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := New(WithMember(member))
				dao.Propose(member, &testPropDef{validationErr: errValidation})
				return dao
			},
			proposalID:   1,
			status:       StatusFailed,
			statusReason: errValidation.Error(),
			err:          errValidation,
		},
		{
			name: "execution error",
			setup: func() *AdminDAO {
				members := []std.Address{
					"g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
					"g1w4ek2u3jta047h6lta047h6lta047h6l9huexc",
					"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				}
				dao := New(WithMember(members[0]), WithMember(members[1]), WithMember(members[2]))
				p, _ := dao.Propose(members[0], &testPropDef{executionErr: errExecution})
				p.record.AddVote(members[0], ChoiceYes)
				p.record.AddVote(members[1], ChoiceYes)
				return dao
			},
			proposalID:   1,
			status:       StatusFailed,
			statusReason: errExecution.Error(),
			err:          errExecution,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dao := tc.setup()

			err := dao.Execute(tc.proposalID)

			if tc.err != nil {
				urequire.ErrorIs(t, err, tc.err)
				return
			}

			urequire.NoError(t, err)

			_, found := dao.GetActiveProposal(tc.proposalID)
			urequire.False(t, found, "proposal should not be active")

			p, found := dao.GetFinishedProposal(tc.proposalID)
			urequire.True(t, found, "proposal not found")
			uassert.Equal(t, string(p.Status()), string(tc.status))
			uassert.Equal(t, string(p.StatusReason()), string(tc.statusReason))
		})
	}
}

type testPropDef struct {
	votingPeriod                time.Duration
	validationErr, executionErr error
}

func (testPropDef) Title() string                 { return "" }
func (testPropDef) Body() string                  { return "" }
func (testPropDef) Quorum() float64               { return 0 }
func (d testPropDef) VotingPeriod() time.Duration { return d.votingPeriod }
func (d testPropDef) Validate() error             { return d.validationErr }
func (d testPropDef) Execute() error              { return d.executionErr }
