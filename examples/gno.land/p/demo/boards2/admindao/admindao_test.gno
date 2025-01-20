package admindao

import (
	"std"
	"testing"
	"time"

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

type testPropDef struct {
	votingPeriod time.Duration
	quorum       float64
}

func (d testPropDef) Title() string               { return "" }
func (d testPropDef) Body() string                { return "" }
func (d testPropDef) VotingPeriod() time.Duration { return 0 }
func (d testPropDef) Quorum() float64             { return 0 }
func (d testPropDef) Validate() error             { return nil }
func (d testPropDef) Execute() error              { return nil }
