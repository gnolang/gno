package commondao_test

import (
	"testing"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"

	"gno.land/p/nt/commondao"
)

func TestVotingRecordDefaults(t *testing.T) {
	var (
		record commondao.VotingRecord
		user   = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	)

	uassert.Equal(t, record.Size(), 0)
	uassert.Equal(t, record.VoteCount(commondao.ChoiceYes), 0)
	uassert.Equal(t, record.VoteCount(commondao.ChoiceNo), 0)
	uassert.Equal(t, record.VoteCount(commondao.ChoiceAbstain), 0)
	uassert.False(t, record.HasVoted(user))
}

func TestVotingRecordAddVote(t *testing.T) {
	cases := []struct {
		name                            string
		setup                           func(*commondao.VotingRecord)
		votes                           []commondao.Vote
		yesCount, noCount, abstainCount int
		updated                         bool
	}{
		{
			name: "single vote",
			votes: []commondao.Vote{
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				},
			},
			yesCount: 1,
		},
		{
			name: "multiple votes",
			votes: []commondao.Vote{
				{
					Address: "g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt",
					Choice:  commondao.ChoiceNo,
				},
				{
					Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk",
					Choice:  commondao.ChoiceYes,
				},
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceNo,
				},
				{
					Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
					Choice:  commondao.ChoiceAbstain,
				},
			},
			yesCount:     1,
			noCount:      2,
			abstainCount: 1,
		},
		{
			name: "vote exists",
			votes: []commondao.Vote{
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				},
			},
			setup: func(r *commondao.VotingRecord) {
				r.AddVote(commondao.Vote{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceAbstain,
				})
			},
			yesCount:     1,
			abstainCount: 0,
			updated:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				record  commondao.VotingRecord
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
			record.Iterate(0, record.Size(), false, func(v commondao.Vote) bool {
				uassert.Equal(t, v.Address, tc.votes[i].Address)
				uassert.Equal(t, string(v.Choice), string(tc.votes[i].Choice))
				uassert.True(t, record.HasVoted(v.Address))

				i++
				return false
			})

			uassert.Equal(t, record.VoteCount(commondao.ChoiceYes), tc.yesCount, "expect YES vote count to match")
			uassert.Equal(t, record.VoteCount(commondao.ChoiceNo), tc.noCount, "expect NO vote count to match")
			uassert.Equal(t, record.VoteCount(commondao.ChoiceAbstain), tc.abstainCount, "expect ABSTAIN vote count to match")
		})
	}
}

func TestFindMostVotedChoice(t *testing.T) {
	cases := []struct {
		name   string
		setup  func(*commondao.VotingRecord)
		choice commondao.VoteChoice
	}{
		{
			name:   "no votes",
			choice: commondao.ChoiceNone,
		},
		{
			name: "one vote",
			setup: func(r *commondao.VotingRecord) {
				r.AddVote(commondao.Vote{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				})
			},
			choice: commondao.ChoiceYes,
		},
		{
			name: "multiple votes",
			setup: func(r *commondao.VotingRecord) {
				r.AddVote(commondao.Vote{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceNo,
				})
				r.AddVote(commondao.Vote{
					Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
					Choice:  commondao.ChoiceNo,
				})
			},
			choice: commondao.ChoiceNo,
		},
		{
			name: "tie",
			setup: func(r *commondao.VotingRecord) {
				r.AddVote(commondao.Vote{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk",
					Choice:  commondao.ChoiceNo,
				})
			},
			choice: commondao.ChoiceNone,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var record commondao.VotingRecord

			if tc.setup != nil {
				tc.setup(&record)
			}

			choice := commondao.FindMostVotedChoice(record.Readonly())

			uassert.Equal(t, string(choice), string(tc.choice))
		})
	}
}

func TestSelectChoiceByAbsoluteMajority(t *testing.T) {
	cases := []struct {
		name         string
		setup        func(*commondao.VotingRecord)
		choice       commondao.VoteChoice
		membersCount int
		success      bool
	}{
		{
			name:         "no votes",
			choice:       commondao.ChoiceNone,
			membersCount: 3,
			success:      false,
		},
		{
			name: "majority",
			setup: func(r *commondao.VotingRecord) {
				r.AddVote(commondao.Vote{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
					Choice:  commondao.ChoiceNo,
				})
			},
			choice:       commondao.ChoiceYes,
			membersCount: 3,
			success:      true,
		},
		{
			name: "no majority",
			setup: func(r *commondao.VotingRecord) {
				r.AddVote(commondao.Vote{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk",
					Choice:  commondao.ChoiceNo,
				})
			},
			choice:       "",
			membersCount: 3,
			success:      false,
		},
		{
			name: "majority with abstain vote",
			setup: func(r *commondao.VotingRecord) {
				r.AddVote(commondao.Vote{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
					Choice:  commondao.ChoiceAbstain,
				})
			},
			choice:       commondao.ChoiceYes,
			membersCount: 3,
			success:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var record commondao.VotingRecord

			if tc.setup != nil {
				tc.setup(&record)
			}

			choice, success := commondao.SelectChoiceByAbsoluteMajority(record.Readonly(), tc.membersCount)

			uassert.Equal(t, string(tc.choice), string(choice), "choice")
			uassert.Equal(t, tc.success, success, "success")
		})
	}
}

func TestSelectChoiceBySuperMajority(t *testing.T) {
	cases := []struct {
		name         string
		setup        func(*commondao.VotingRecord)
		choice       commondao.VoteChoice
		membersCount int
		success      bool
	}{
		{
			name:         "no votes",
			choice:       commondao.ChoiceNone,
			membersCount: 3,
			success:      false,
		},
		{
			name: "majority",
			setup: func(r *commondao.VotingRecord) {
				r.AddVote(commondao.Vote{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
					Choice:  commondao.ChoiceNo,
				})
			},
			choice:       commondao.ChoiceYes,
			membersCount: 3,
			success:      true,
		},
		{
			name: "no majority",
			setup: func(r *commondao.VotingRecord) {
				r.AddVote(commondao.Vote{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk",
					Choice:  commondao.ChoiceNo,
				})
			},
			choice:       "",
			membersCount: 3,
			success:      false,
		},
		{
			name: "majority with abstain vote",
			setup: func(r *commondao.VotingRecord) {
				r.AddVote(commondao.Vote{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
					Choice:  commondao.ChoiceAbstain,
				})
			},
			choice:       commondao.ChoiceYes,
			membersCount: 3,
			success:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var record commondao.VotingRecord

			if tc.setup != nil {
				tc.setup(&record)
			}

			choice, success := commondao.SelectChoiceBySuperMajority(record.Readonly(), tc.membersCount)

			uassert.Equal(t, string(tc.choice), string(choice), "choice")
			uassert.Equal(t, tc.success, success, "success")
		})
	}
}

func TestSelectChoiceByPlurality(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(*commondao.VotingRecord)
		choice  commondao.VoteChoice
		success bool
	}{
		{
			name:    "no votes",
			choice:  commondao.ChoiceNone,
			success: false,
		},
		{
			name: "plurality",
			setup: func(r *commondao.VotingRecord) {
				r.AddVote(commondao.Vote{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
					Choice:  commondao.ChoiceNo,
				})
			},
			choice:  commondao.ChoiceYes,
			success: true,
		},
		{
			name: "no plurality",
			setup: func(r *commondao.VotingRecord) {
				r.AddVote(commondao.Vote{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk",
					Choice:  commondao.ChoiceNo,
				})
			},
			choice:  "",
			success: false,
		},
		{
			name: "plurality with abstain vote",
			setup: func(r *commondao.VotingRecord) {
				r.AddVote(commondao.Vote{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g12chzmwxw8sezcxe9h2csp0tck76r4ptwdlyyqk",
					Choice:  commondao.ChoiceYes,
				})
				r.AddVote(commondao.Vote{
					Address: "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5",
					Choice:  commondao.ChoiceNo,
				})
				r.AddVote(commondao.Vote{
					Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
					Choice:  commondao.ChoiceAbstain,
				})
			},
			choice:  commondao.ChoiceYes,
			success: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var record commondao.VotingRecord

			if tc.setup != nil {
				tc.setup(&record)
			}

			choice, success := commondao.SelectChoiceByPlurality(record.Readonly())

			uassert.Equal(t, string(tc.choice), string(choice), "choice")
			uassert.Equal(t, tc.success, success, "success")
		})
	}
}
