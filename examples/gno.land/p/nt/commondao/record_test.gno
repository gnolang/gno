package commondao

import (
	"chain"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestVotingRecordDefaults(t *testing.T) {
	var (
		record VotingRecord
		user   = chain.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	)

	uassert.Equal(t, record.Size(), 0)
	uassert.Equal(t, record.VoteCount(ChoiceYes), 0)
	uassert.Equal(t, record.VoteCount(ChoiceNo), 0)
	uassert.Equal(t, record.VoteCount(ChoiceAbstain), 0)
	uassert.False(t, record.HasVoted(user))
}

func TestVotingRecordAddVote(t *testing.T) {
	cases := []struct {
		name                            string
		setup                           func(*VotingRecord)
		votes                           []Vote
		yesCount, noCount, abstainCount int
		updated                         bool
	}{
		{
			name: "single vote",
			votes: []Vote{
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  ChoiceYes,
				},
			},
			yesCount: 1,
		},
		{
			name: "multiple votes",
			votes: []Vote{
				{
					Address: "g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt",
					Choice:  ChoiceNo,
				},
				{
					Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk",
					Choice:  ChoiceYes,
				},
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  ChoiceNo,
				},
				{
					Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
					Choice:  ChoiceAbstain,
				},
			},
			yesCount:     1,
			noCount:      2,
			abstainCount: 1,
		},
		{
			name: "vote exists",
			votes: []Vote{
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  ChoiceYes,
				},
			},
			setup: func(r *VotingRecord) {
				r.AddVote(Vote{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceAbstain})
			},
			yesCount:     1,
			abstainCount: 0,
			updated:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				record  VotingRecord
				updated bool
			)

			if tc.setup != nil {
				tc.setup(&record)
			}

			for _, v := range tc.votes {
				updated = updated || record.AddVote(v)
			}

			urequire.Equal(t, updated, tc.updated, "expect vote to be updated")
			urequire.Equal(t, record.Size(), len(tc.votes), "expect record size to match")

			var i int
			record.Iterate(0, record.Size(), false, func(v Vote) bool {
				uassert.Equal(t, v.Address, tc.votes[i].Address)
				uassert.Equal(t, string(v.Choice), string(tc.votes[i].Choice))
				uassert.True(t, record.HasVoted(v.Address))

				i++
				return false
			})

			uassert.Equal(t, record.VoteCount(ChoiceYes), tc.yesCount, "expect YES vote count to match")
			uassert.Equal(t, record.VoteCount(ChoiceNo), tc.noCount, "expect NO vote count to match")
			uassert.Equal(t, record.VoteCount(ChoiceAbstain), tc.abstainCount, "expect ABSTAIN vote count to match")
		})
	}
}

func TestFindMostVotedChoice(t *testing.T) {
	cases := []struct {
		name   string
		setup  func(*VotingRecord)
		choice VoteChoice
	}{
		{
			name:   "no votes",
			choice: ChoiceNone,
		},
		{
			name: "one vote",
			setup: func(r *VotingRecord) {
				r.AddVote(Vote{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes})
			},
			choice: ChoiceYes,
		},
		{
			name: "multiple votes",
			setup: func(r *VotingRecord) {
				r.AddVote(Vote{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceNo})
				r.AddVote(Vote{Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", Choice: ChoiceNo})
			},
			choice: ChoiceNo,
		},
		{
			name: "tie",
			setup: func(r *VotingRecord) {
				r.AddVote(Vote{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk", Choice: ChoiceNo})
			},
			choice: ChoiceNone,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var record VotingRecord

			if tc.setup != nil {
				tc.setup(&record)
			}

			choice := FindMostVotedChoice(record.Readonly())

			uassert.Equal(t, string(choice), string(tc.choice))
		})
	}
}

func TestSelectChoiceByAbsoluteMajority(t *testing.T) {
	cases := []struct {
		name         string
		setup        func(*VotingRecord)
		choice       VoteChoice
		membersCount int
		success      bool
	}{
		{
			name:         "no votes",
			choice:       ChoiceNone,
			membersCount: 3,
			success:      false,
		},
		{
			name: "majority",
			setup: func(r *VotingRecord) {
				r.AddVote(Vote{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", Choice: ChoiceNo})
			},
			choice:       ChoiceYes,
			membersCount: 3,
			success:      true,
		},
		{
			name: "no majority",
			setup: func(r *VotingRecord) {
				r.AddVote(Vote{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk", Choice: ChoiceNo})
			},
			choice:       "",
			membersCount: 3,
			success:      false,
		},
		{
			name: "majority with abstain vote",
			setup: func(r *VotingRecord) {
				r.AddVote(Vote{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", Choice: ChoiceAbstain})
			},
			choice:       ChoiceYes,
			membersCount: 3,
			success:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var record VotingRecord

			if tc.setup != nil {
				tc.setup(&record)
			}

			choice, success := SelectChoiceByAbsoluteMajority(record.Readonly(), tc.membersCount)

			uassert.Equal(t, string(tc.choice), string(choice), "choice")
			uassert.Equal(t, tc.success, success, "success")
		})
	}
}

func TestSelectChoiceBySuperMajority(t *testing.T) {
	cases := []struct {
		name         string
		setup        func(*VotingRecord)
		choice       VoteChoice
		membersCount int
		success      bool
	}{
		{
			name:         "no votes",
			choice:       ChoiceNone,
			membersCount: 3,
			success:      false,
		},
		{
			name: "majority",
			setup: func(r *VotingRecord) {
				r.AddVote(Vote{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", Choice: ChoiceNo})
			},
			choice:       ChoiceYes,
			membersCount: 3,
			success:      true,
		},
		{
			name: "no majority",
			setup: func(r *VotingRecord) {
				r.AddVote(Vote{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk", Choice: ChoiceNo})
			},
			choice:       "",
			membersCount: 3,
			success:      false,
		},
		{
			name: "majority with abstain vote",
			setup: func(r *VotingRecord) {
				r.AddVote(Vote{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", Choice: ChoiceAbstain})
			},
			choice:       ChoiceYes,
			membersCount: 3,
			success:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var record VotingRecord

			if tc.setup != nil {
				tc.setup(&record)
			}

			choice, success := SelectChoiceBySuperMajority(record.Readonly(), tc.membersCount)

			uassert.Equal(t, string(tc.choice), string(choice), "choice")
			uassert.Equal(t, tc.success, success, "success")
		})
	}
}

func TestSelectChoiceByPlurality(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(*VotingRecord)
		choice  VoteChoice
		success bool
	}{
		{
			name:    "no votes",
			choice:  ChoiceNone,
			success: false,
		},
		{
			name: "plurality",
			setup: func(r *VotingRecord) {
				r.AddVote(Vote{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", Choice: ChoiceNo})
			},
			choice:  ChoiceYes,
			success: true,
		},
		{
			name: "no plurality",
			setup: func(r *VotingRecord) {
				r.AddVote(Vote{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk", Choice: ChoiceNo})
			},
			choice:  "",
			success: false,
		},
		{
			name: "plurality with abstain vote",
			setup: func(r *VotingRecord) {
				r.AddVote(Vote{Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk", Choice: ChoiceYes})
				r.AddVote(Vote{Address: "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5", Choice: ChoiceNo})
				r.AddVote(Vote{Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", Choice: ChoiceAbstain})
			},
			choice:  ChoiceYes,
			success: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var record VotingRecord

			if tc.setup != nil {
				tc.setup(&record)
			}

			choice, success := SelectChoiceByPlurality(record.Readonly())

			uassert.Equal(t, string(tc.choice), string(choice), "choice")
			uassert.Equal(t, tc.success, success, "success")
		})
	}
}
