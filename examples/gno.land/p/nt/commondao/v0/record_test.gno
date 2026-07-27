package commondao_test

import (
	"testing"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/uassert/v0"

	"gno.land/p/nt/commondao/v0"
)

func TestVotingRecordDefaults(t *testing.T) {
	var record commondao.VotingRecord

	uassert.Equal(t, 0, record.Size())
	uassert.Equal(t, 0, record.VoteCount(commondao.ChoiceYes))
	uassert.False(t, record.HasVoted(memberA))

	_, found := record.GetVote(memberA)
	uassert.False(t, found)
}

func TestVotingRecordAddVote(t *testing.T) {
	var record commondao.VotingRecord

	updated := record.AddVote(commondao.Vote{
		Address: memberA,
		Choice:  commondao.ChoiceYes,
		Reason:  "foo",
	})
	uassert.False(t, updated, "expect first vote to be an insert")
	uassert.Equal(t, 1, record.Size())
	uassert.Equal(t, 1, record.VoteCount(commondao.ChoiceYes))
	uassert.True(t, record.HasVoted(memberA))

	v, found := record.GetVote(memberA)
	uassert.True(t, found)
	uassert.Equal(t, string(commondao.ChoiceYes), string(v.Choice))
	uassert.Equal(t, "foo", v.Reason)

	// Voting again overwrites the previous vote and updates the counters
	updated = record.AddVote(commondao.Vote{
		Address: memberA,
		Choice:  commondao.ChoiceNo,
	})
	uassert.True(t, updated, "expect second vote to be an update")
	uassert.Equal(t, 1, record.Size())
	uassert.Equal(t, 0, record.VoteCount(commondao.ChoiceYes))
	uassert.Equal(t, 1, record.VoteCount(commondao.ChoiceNo))

	record.AddVote(commondao.Vote{Address: memberB, Choice: commondao.ChoiceNo})
	uassert.Equal(t, 2, record.Size())
	uassert.Equal(t, 2, record.VoteCount(commondao.ChoiceNo))

	count := 0
	record.Iterate(0, record.Size(), false, func(commondao.Vote) bool {
		count++
		return false
	})
	uassert.Equal(t, 2, count)
}

func TestTallyDefault(t *testing.T) {
	cases := []struct {
		name       string
		electorate []address
		yes        []address
		no         []address
		abstain    []address
		threshold  commondao.Threshold
		want       commondao.Outcome
	}{
		{
			name: "empty electorate stays pending",
			want: commondao.OutcomePending,
		},
		{
			name:       "no votes stays pending",
			electorate: []address{memberA, memberB, memberC},
			want:       commondao.OutcomePending,
		},
		{
			name:       "all abstain stays pending",
			electorate: []address{memberA, memberB, memberC},
			abstain:    []address{memberA, memberB, memberC},
			want:       commondao.OutcomePending,
		},
		{
			name:       "supermajority boundary passes",
			electorate: []address{memberA, memberB, memberC},
			yes:        []address{memberA, memberB}, // 3*2 >= 2*3
			want:       commondao.OutcomePassed,
		},
		{
			name:       "below supermajority stays pending",
			electorate: []address{memberA, memberB, memberC},
			yes:        []address{memberA},
			no:         []address{memberB},
			want:       commondao.OutcomePending,
		},
		{
			name:       "abstains shrink the denominator for passing",
			electorate: []address{memberA, memberB, memberC},
			yes:        []address{memberA},
			abstain:    []address{memberB, memberC}, // D=1, 3*1 >= 2*1
			want:       commondao.OutcomePassed,
		},
		{
			name:       "silence counts against passage",
			electorate: []address{memberA, memberB, memberC, memberD, memberE, memberF, memberG, memberH, memberI},
			yes:        []address{memberA, memberB, memberC, memberD},
			no:         []address{memberE}, // 4Y/1N/4 silent: 12 < 18
			want:       commondao.OutcomePending,
		},
		{
			name:       "NO majority boundary is strict",
			electorate: []address{memberA, memberB, memberC, memberD},
			no:         []address{memberA, memberB}, // 2*2 == 4: not dismissed
			want:       commondao.OutcomePending,
		},
		{
			name:       "NO majority dismisses",
			electorate: []address{memberA, memberB, memberC, memberD},
			no:         []address{memberA, memberB, memberC}, // 2*3 > 4
			want:       commondao.OutcomeDismissed,
		},
		{
			name:       "abstains shrink the denominator for dismissal",
			electorate: []address{memberA, memberB, memberC, memberD, memberE},
			no:         []address{memberA, memberB},
			abstain:    []address{memberC, memberD}, // D=3, 2*2 > 3
			want:       commondao.OutcomeDismissed,
		},
		{
			name:       "simple majority boundary does not pass",
			electorate: []address{memberA, memberB, memberC, memberD},
			yes:        []address{memberA, memberB}, // 2*2 == 4
			threshold:  commondao.ThresholdSimpleMajority,
			want:       commondao.OutcomePending,
		},
		{
			name:       "simple majority passes",
			electorate: []address{memberA, memberB, memberC, memberD},
			yes:        []address{memberA, memberB, memberC}, // 2*3 > 4
			threshold:  commondao.ThresholdSimpleMajority,
			want:       commondao.OutcomePassed,
		},
		{
			name:       "simple majority minimal boundary passes",
			electorate: []address{memberA, memberB, memberC},
			yes:        []address{memberA, memberB}, // 2*2 == D+1
			threshold:  commondao.ThresholdSimpleMajority,
			want:       commondao.OutcomePassed,
		},
		{
			name:       "votes from outside the electorate are not counted",
			electorate: []address{memberA, memberB},
			yes:        []address{memberC, memberD}, // strangers
			want:       commondao.OutcomePending,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				electorate addrset.Set
				record     commondao.VotingRecord
			)
			for _, m := range tc.electorate {
				electorate.Add(m)
			}
			for _, m := range tc.yes {
				record.AddVote(commondao.Vote{Address: m, Choice: commondao.ChoiceYes})
			}
			for _, m := range tc.no {
				record.AddVote(commondao.Vote{Address: m, Choice: commondao.ChoiceNo})
			}
			for _, m := range tc.abstain {
				record.AddVote(commondao.Vote{Address: m, Choice: commondao.ChoiceAbstain})
			}

			got := commondao.TallyDefault(record.Readonly(), electorate.Readonly(), tc.threshold)

			uassert.Equal(t, int(tc.want), int(got))
		})
	}
}
