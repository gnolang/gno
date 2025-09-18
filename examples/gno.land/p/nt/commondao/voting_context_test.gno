package commondao_test

import (
	"testing"

	"gno.land/p/nt/urequire"

	"gno.land/p/nt/commondao"
)

func TestCollectVotes(t *testing.T) {
	cases := []struct {
		name   string
		setup  func() commondao.MemberStorage
		votes  []commondao.Vote
		groups []string
		error  string
	}{
		{
			name: "one group",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorageWithGrouping()
				one, _ := s.Grouping().Add("one")
				one.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				one.Members().Add("g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt")
				return s
			},
			groups: []string{"one"},
			votes: []commondao.Vote{
				{
					Address: "g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt",
					Choice:  commondao.ChoiceNo,
				},
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				},
			},
		},
		{
			name: "two groups",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorageWithGrouping()
				one, _ := s.Grouping().Add("one")
				one.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				one.Members().Add("g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt")
				two, _ := s.Grouping().Add("two")
				two.Members().Add("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
				return s
			},
			groups: []string{"one", "two"},
			votes: []commondao.Vote{
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				},
				{
					Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
					Choice:  commondao.ChoiceNo,
				},
			},
		},
		{
			name:  "no group names",
			setup: func() commondao.MemberStorage { return commondao.NewMemberStorageWithGrouping() },
			error: "one or more group names are required to collect votes",
		},
		{
			name:   "member group not found",
			setup:  func() commondao.MemberStorage { return commondao.NewMemberStorageWithGrouping() },
			groups: []string{"foo"},
			error:  "member group not found: foo",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var r commondao.VotingRecord
			for _, v := range tc.votes {
				r.AddVote(v)
			}

			storage := tc.setup()
			ctx := commondao.MustNewVotingContext(&r, storage)

			// Act
			votes, err := commondao.CollectVotes(ctx, tc.groups...)

			// Assert
			if tc.error != "" {
				urequire.ErrorContains(t, err, tc.error)
				return
			}

			urequire.NoError(t, err, "unexpected error")
			urequire.Equal(t, len(tc.votes), votes.Size(), "expect number of votes to match")

			var i int
			votes.Iterate("", "", func(addr string, v any) bool {
				want := tc.votes[i]
				got := v.(commondao.Vote)

				urequire.Equal(t, want.Address, got.Address, "expect vote address to match")
				urequire.Equal(t, string(want.Choice), string(got.Choice), "expect vote choice to match")
				urequire.Equal(t, want.Address.String(), addr, "expect vote address to match votes key")

				i++
				return false
			})
		})
	}
}
