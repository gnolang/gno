package commondao_test

import (
	"errors"
	"testing"
	"time"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"

	"gno.land/p/nt/commondao"
)

func TestNew(t *testing.T) {
	cases := []struct {
		name    string
		parent  *commondao.CommonDAO
		members []address
	}{
		{
			name:    "with parent",
			parent:  commondao.New(),
			members: []address{"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},
		},
		{
			name:    "without parent",
			members: []address{"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},
		},
		{
			name: "multiple members",
			members: []address{
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
			options := []commondao.Option{commondao.WithParent(tc.parent)}
			for _, m := range tc.members {
				options = append(options, commondao.WithMember(m))
			}

			dao := commondao.New(options...)

			if tc.parent == nil {
				uassert.Equal(t, nil, dao.Parent())
			} else {
				uassert.NotEqual(t, nil, dao.Parent())
			}

			uassert.False(t, dao.IsDeleted(), "expect DAO not to be soft deleted by default")
			urequire.Equal(t, membersCount, dao.Members().Size(), "dao members")

			var i int
			dao.Members().IterateByOffset(0, membersCount, func(addr address) bool {
				uassert.Equal(t, tc.members[i], addr)
				i++
				return false
			})
		})
	}
}

func TestCommonDAOMembersAdd(t *testing.T) {
	member := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	dao := commondao.New(commondao.WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"))

	added := dao.Members().Add(member)
	urequire.True(t, added)

	uassert.Equal(t, 2, dao.Members().Size())
	uassert.True(t, dao.Members().Has(member))

	added = dao.Members().Add(member)
	urequire.False(t, added)
}

func TestCommonDAOMembersRemove(t *testing.T) {
	member := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	dao := commondao.New(commondao.WithMember(member))

	removed := dao.Members().Remove(member)
	urequire.True(t, removed)

	removed = dao.Members().Remove(member)
	urequire.False(t, removed)
}

func TestCommonDAOMembersHas(t *testing.T) {
	cases := []struct {
		name   string
		member address
		dao    *commondao.CommonDAO
		want   bool
	}{
		{
			name:   "member",
			member: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			dao:    commondao.New(commondao.WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")),
			want:   true,
		},
		{
			name:   "not a dao member",
			member: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			dao:    commondao.New(commondao.WithMember("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc")),
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
		setup   func() *commondao.CommonDAO
		creator address
		def     commondao.ProposalDefinition
		err     error
	}{
		{
			name:    "success",
			setup:   func() *commondao.CommonDAO { return commondao.New() },
			creator: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			def:     testPropDef{},
		},
		{
			name:  "nil definition",
			setup: func() *commondao.CommonDAO { return commondao.New() },
			err:   commondao.ErrProposalDefinitionRequired,
		},
		{
			name:  "invalid creator address",
			setup: func() *commondao.CommonDAO { return commondao.New() },
			def:   testPropDef{},
			err:   commondao.ErrInvalidCreatorAddress,
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

func TestCommonDAOWithdraw(t *testing.T) {
	cases := []struct {
		name       string
		proposalID uint64
		setup      func() *commondao.CommonDAO
		err        error
	}{
		{
			name:       "success",
			proposalID: 1,
			setup: func() *commondao.CommonDAO {
				member := address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := commondao.New(commondao.WithMember(member))
				dao.Propose(member, testPropDef{votingPeriod: time.Hour})
				return dao
			},
		},
		{
			name:       "proposal not found",
			proposalID: 404,
			setup: func() *commondao.CommonDAO {
				return commondao.New(commondao.WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"))
			},
			err: commondao.ErrProposalNotFound,
		},
		{
			name:       "withdrawal not allowed",
			proposalID: 1,
			setup: func() *commondao.CommonDAO {
				member := address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := commondao.New(commondao.WithMember(member))
				p, _ := dao.Propose(member, testPropDef{votingPeriod: time.Hour})
				dao.Vote(member, p.ID(), commondao.ChoiceYes, "")
				return dao
			},
			err: commondao.ErrWithdrawalNotAllowed,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dao := tc.setup()

			err := dao.Withdraw(tc.proposalID)

			if tc.err != nil {
				urequire.ErrorIs(t, err, tc.err)
				return
			}

			urequire.NoError(t, err)
		})
	}
}

func TestCommonDAOVote(t *testing.T) {
	cases := []struct {
		name       string
		setup      func() *commondao.CommonDAO
		member     address
		choice     commondao.VoteChoice
		proposalID uint64
		err        error
	}{
		{
			name: "success",
			setup: func() *commondao.CommonDAO {
				member := address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := commondao.New(commondao.WithMember(member))
				dao.Propose(member, testPropDef{votingPeriod: time.Hour})
				return dao
			},
			member:     "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			choice:     commondao.ChoiceYes,
			proposalID: 1,
		},
		{
			name: "success with custom vote choice",
			setup: func() *commondao.CommonDAO {
				member := address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := commondao.New(commondao.WithMember(member))
				dao.Propose(member, testPropDef{
					votingPeriod: time.Hour,
					voteChoices:  []commondao.VoteChoice{"FOO", "BAR"},
				})
				return dao
			},
			member:     "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			choice:     commondao.VoteChoice("BAR"),
			proposalID: 1,
		},
		{
			name: "success with deadline check disabled",
			setup: func() *commondao.CommonDAO {
				member := address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := commondao.New(
					commondao.WithMember(member),
					commondao.DisableVotingDeadlineCheck(),
				)
				dao.Propose(member, testPropDef{})
				return dao
			},
			member:     "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			choice:     commondao.ChoiceYes,
			proposalID: 1,
		},
		{
			name: "invalid vote choice",
			setup: func() *commondao.CommonDAO {
				member := address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
				dao := commondao.New(commondao.WithMember(member))
				dao.Propose(member, testPropDef{votingPeriod: time.Hour})
				return dao
			},
			member:     "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			choice:     commondao.VoteChoice("invalid"),
			proposalID: 1,
			err:        commondao.ErrInvalidVoteChoice,
		},
		{
			name:   "not a member",
			setup:  func() *commondao.CommonDAO { return commondao.New() },
			member: "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			choice: commondao.ChoiceAbstain,
			err:    commondao.ErrNotMember,
		},
		{
			name: "proposal not found",
			setup: func() *commondao.CommonDAO {
				return commondao.New(commondao.WithMember("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"))
			},
			member:     "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn",
			choice:     commondao.ChoiceAbstain,
			proposalID: 42,
			err:        commondao.ErrProposalNotFound,
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

func TestCommonDAOExecute(t *testing.T) {
	errTest := errors.New("test")
	member := address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
	cases := []struct {
		name         string
		setup        func() *commondao.CommonDAO
		proposalID   uint64
		status       commondao.ProposalStatus
		statusReason string
		err          error
	}{
		// TODO: Execution success and error are implemented as filetests
		//       This is done because proposal definition's Execute() must be
		//       crossing which is not possible without defining it within a realm.
		// {
		// 	name: "success",
		// 	setup: func() *commondao.CommonDAO {
		// 		dao := commondao.New(commondao.WithMember(member))
		// 		dao.Propose(member, testPropDef{tallyResult: true})
		// 		return dao
		// 	},
		// 	status:     StatusPassed,
		// 	proposalID: 1,
		// },
		// {
		// 	name: "execution error",
		// 	setup: func() *commondao.CommonDAO {
		// 		dao := commondao.New(commondao.WithMember(member))
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
			setup:      func() *commondao.CommonDAO { return commondao.New() },
			proposalID: 1,
			err:        commondao.ErrProposalNotFound,
		},
		{
			name: "execution not allowed",
			setup: func() *commondao.CommonDAO {
				dao := commondao.New(commondao.WithMember(member))
				p, _ := dao.Propose(member, testPropDef{tallyResult: false})
				p.Tally(dao.Members())
				return dao
			},
			proposalID: 1,
			err:        commondao.ErrExecutionNotAllowed,
		},
		{
			name: "voting deadline not met",
			setup: func() *commondao.CommonDAO {
				dao := commondao.New(commondao.WithMember(member))
				dao.Propose(member, testPropDef{votingPeriod: time.Minute * 5})
				return dao
			},
			proposalID: 1,
			err:        commondao.ErrVotingDeadlineNotMet,
		},
		{
			name: "validation error",
			setup: func() *commondao.CommonDAO {
				dao := commondao.New(commondao.WithMember(member))
				dao.Propose(member, testPropDef{
					validationErr: errTest,
					tallyResult:   true,
				})
				return dao
			},
			proposalID:   1,
			status:       commondao.StatusFailed,
			statusReason: errTest.Error(),
		},
		{
			name: "tally error",
			setup: func() *commondao.CommonDAO {
				dao := commondao.New(commondao.WithMember(member))
				dao.Propose(member, testPropDef{tallyErr: errTest})
				return dao
			},
			proposalID:   1,
			status:       commondao.StatusFailed,
			statusReason: errTest.Error(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dao := tc.setup()

			err := dao.Execute(tc.proposalID)

			if tc.err != nil {
				urequire.Error(t, err, "expected an error")
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
