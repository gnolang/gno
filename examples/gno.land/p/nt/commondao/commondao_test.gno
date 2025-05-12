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
			name:    "success",
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

			found := dao.ActiveProposals().Has(p.ID())
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
			name: "success",
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
			name: "success with custom vote choice",
			setup: func() *CommonDAO {
				member := std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := New(WithMember(member))
				dao.Propose(member, testPropDef{
					voteChoices: []VoteChoice{"FOO", "BAR"},
				})
				return dao
			},
			member:     "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			choice:     VoteChoice("BAR"),
			proposalID: 1,
		},
		{
			name: "invalid vote choice",
			setup: func() *CommonDAO {
				member := std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := New(WithMember(member))
				dao.Propose(member, testPropDef{})
				return dao
			},
			member:     "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			choice:     VoteChoice("invalid"),
			proposalID: 1,
			err:        ErrInvalidVoteChoice,
		},
		{
			name:   "not a member",
			setup:  func() *CommonDAO { return New() },
			member: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			choice: ChoiceAbstain,
			err:    ErrNotMember,
		},
		{
			name: "proposal not found",
			setup: func() *CommonDAO {
				return New(WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"))
			},
			member:     "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			choice:     ChoiceAbstain,
			proposalID: 42,
			err:        ErrProposalNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dao := tc.setup()

			err := dao.Vote(tc.member, tc.proposalID, tc.choice, "")

			if tc.err != nil {
				urequire.ErrorIs(t, err, tc.err)
				return
			}

			urequire.NoError(t, err)

			p := dao.ActiveProposals().Get(tc.proposalID)
			urequire.NotEqual(t, nil, p, "proposal not found")

			record := p.VotingRecord()
			uassert.True(t, record.HasVoted(tc.member))
			uassert.Equal(t, record.VoteCount(tc.choice), 1)
		})
	}
}

func TestCommonDAOTally(t *testing.T) {
	errTest := errors.New("test")
	member := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	cases := []struct {
		name   string
		setup  func(*CommonDAO) (proposalID uint64)
		passes bool
		err    error
	}{
		{
			name: "pass",
			setup: func(dao *CommonDAO) uint64 {
				return dao.MustPropose(member, testPropDef{tallyResult: true}).ID()
			},
			passes: true,
		},
		{
			name: "fail to pass",
			setup: func(dao *CommonDAO) uint64 {
				return dao.MustPropose(member, testPropDef{tallyResult: false}).ID()
			},
			passes: false,
		},
		{
			name:  "proposal not found",
			setup: func(*CommonDAO) uint64 { return 404 },
			err:   ErrProposalNotFound,
		},
		{
			name: "proposal status not active",
			setup: func(dao *CommonDAO) uint64 {
				p := dao.MustPropose(member, testPropDef{})
				p.status = StatusPassed
				return p.ID()
			},
			err: ErrStatusIsNotActive,
		},
		{
			name: "proposal failed error",
			setup: func(dao *CommonDAO) uint64 {
				return dao.MustPropose(member, testPropDef{tallyErr: ErrProposalFailed}).ID()
			},
			passes: false,
		},
		{
			name: "error",
			setup: func(dao *CommonDAO) uint64 {
				return dao.MustPropose(member, testPropDef{tallyErr: errTest}).ID()
			},
			err: errTest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dao := New(WithMember(member))
			proposalID := tc.setup(dao)

			passes, err := dao.Tally(proposalID)

			if tc.err != nil {
				uassert.ErrorIs(t, err, tc.err, "expect an error")
				uassert.False(t, passes, "expect tally to fail")
				return
			}

			uassert.NoError(t, err, "expect no error")
			uassert.Equal(t, tc.passes, passes, "expect tally success value to match")
		})
	}
}

func TestCommonDAOExecute(t *testing.T) {
	errTest := errors.New("test")
	member := std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
	cases := []struct {
		name         string
		setup        func() *CommonDAO
		proposalID   uint64
		status       ProposalStatus
		statusReason string
		err          error
	}{
		// TODO: Execution success and error are implemented as filetests
		//       This is done because proposal definition's Execute() must be
		//       crossing which is not possible without defining it within a realm.
		// {
		// 	name: "success",
		// 	setup: func() *CommonDAO {
		// 		dao := New(WithMember(member))
		// 		dao.Propose(member, testPropDef{tallyResult: true})
		// 		return dao
		// 	},
		// 	status:     StatusPassed,
		// 	proposalID: 1,
		// },
		// {
		// 	name: "execution error",
		// 	setup: func() *CommonDAO {
		// 		dao := New(WithMember(member))
		// 		dao.Propose(member, testPropDef{
		// 			tallyResult:  true,
		// 			executionErr: errTest,
		// 		})
		// 		return dao
		// 	},
		// 	proposalID:   1,
		// 	status:       StatusFailed,
		// 	statusReason: errTest.Error(),
		// },
		{
			name:       "proposal not found",
			setup:      func() *CommonDAO { return New() },
			proposalID: 1,
			err:        ErrProposalNotFound,
		},
		{
			name: "proposal not active",
			setup: func() *CommonDAO {
				dao := New(WithMember(member))
				p, _ := dao.Propose(member, testPropDef{})
				p.status = StatusPassed
				return dao
			},
			proposalID: 1,
			err:        ErrStatusIsNotActive,
		},
		{
			name: "voting deadline not met",
			setup: func() *CommonDAO {
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
				dao := New(WithMember(member))
				dao.Propose(member, testPropDef{validationErr: errTest})
				return dao
			},
			proposalID:   1,
			status:       StatusFailed,
			statusReason: errTest.Error(),
		},
		{
			name: "tally error",
			setup: func() *CommonDAO {
				dao := New(WithMember(member))
				dao.Propose(member, testPropDef{tallyErr: errTest})
				return dao
			},
			proposalID:   1,
			status:       StatusFailed,
			statusReason: errTest.Error(),
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

			found := dao.ActiveProposals().Has(tc.proposalID)
			urequire.False(t, found, "proposal should not be active")

			p := dao.FinishedProposals().Get(tc.proposalID)
			urequire.NotEqual(t, nil, p, "proposal must be found")
			uassert.Equal(t, string(p.Status()), string(tc.status), "status must match")
			uassert.Equal(t, string(p.StatusReason()), string(tc.statusReason), "status reason must match")
		})
	}
}
