package admindao

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestVotingRecordDefaults(t *testing.T) {
	var (
		record VotingRecord
		user   = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	)

	uassert.Equal(t, record.Votes(), nil)
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
		err                             error
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
				r.AddVote("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", ChoiceAbstain)
			},
			err: ErrVoteExists,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				err    error
				record VotingRecord
			)

			if tc.setup != nil {
				tc.setup(&record)
			}

			for _, v := range tc.votes {
				err = record.AddVote(v.Address, v.Choice)
				if err != nil {
					break
				}
			}

			if tc.err != nil {
				urequire.ErrorIs(t, err, tc.err)
				return
			}

			urequire.NoError(t, err)
			urequire.Equal(t, len(record.Votes()), len(tc.votes), "unexpected number of votes")
			for i, v := range record.Votes() {
				uassert.Equal(t, v.Address, tc.votes[i].Address)
				uassert.Equal(t, string(v.Choice), string(tc.votes[i].Choice))
				uassert.True(t, record.HasVoted(v.Address))
			}

			uassert.Equal(t, record.VoteCount(ChoiceYes), tc.yesCount)
			uassert.Equal(t, record.VoteCount(ChoiceNo), tc.noCount)
			uassert.Equal(t, record.VoteCount(ChoiceAbstain), tc.abstainCount)
		})
	}
}

func TestVotingRecordGetProvableMajorityChoice(t *testing.T) {
	cases := []struct {
		name   string
		setup  func(*VotingRecord)
		choice VoteChoice
	}{
		{
			name:   "no votes",
			choice: ChoiceAbstain,
		},
		{
			name: "one vote",
			setup: func(r *VotingRecord) {
				r.AddVote("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", ChoiceYes)
			},
			choice: ChoiceYes,
		},
		{
			name: "majority",
			setup: func(r *VotingRecord) {
				r.AddVote("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", ChoiceYes)
				r.AddVote("g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk", ChoiceYes)
				r.AddVote("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", ChoiceNo)
			},
			choice: ChoiceYes,
		},
		{
			name: "invalid because no majority",
			setup: func(r *VotingRecord) {
				r.AddVote("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", ChoiceYes)
				r.AddVote("g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk", ChoiceNo)
			},
			choice: ChoiceNo,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var record VotingRecord

			if tc.setup != nil {
				tc.setup(&record)
			}

			choice := record.GetProvableMajorityChoice()

			uassert.Equal(t, string(choice), string(tc.choice))
		})
	}
}
