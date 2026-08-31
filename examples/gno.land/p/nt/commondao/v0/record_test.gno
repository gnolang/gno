package commondao_test

import (
	"testing"

	"gno.land/p/nt/addrset/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/p/nt/commondao/v0"
)

// mustNewVote builds a reason-less vote or fails the test.
func mustNewVote(t *testing.T, addr address, c commondao.VoteChoice) commondao.Vote {
	t.Helper()
	v, err := commondao.NewVote(addr, c, "")
	urequire.NoError(t, err, "new vote")
	return v
}

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

	vote, err := commondao.NewVote(memberA, commondao.ChoiceYes, "foo")
	urequire.NoError(t, err)
	updated := record.AddVote(vote)
	uassert.False(t, updated, "expect first vote to be an insert")
	uassert.Equal(t, 1, record.Size())
	uassert.Equal(t, 1, record.VoteCount(commondao.ChoiceYes))
	uassert.True(t, record.HasVoted(memberA))

	v, found := record.GetVote(memberA)
	uassert.True(t, found)
	uassert.Equal(t, string(commondao.ChoiceYes), string(v.Choice()))
	uassert.Equal(t, "foo", v.Reason())

	// Voting again overwrites the previous vote and updates the counters
	updated = record.AddVote(mustNewVote(t, memberA, commondao.ChoiceNo))
	uassert.True(t, updated, "expect second vote to be an update")
	uassert.Equal(t, 1, record.Size())
	uassert.Equal(t, 0, record.VoteCount(commondao.ChoiceYes))
	uassert.Equal(t, 1, record.VoteCount(commondao.ChoiceNo))

	record.AddVote(mustNewVote(t, memberB, commondao.ChoiceNo))
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
				record.AddVote(mustNewVote(t, m, commondao.ChoiceYes))
			}
			for _, m := range tc.no {
				record.AddVote(mustNewVote(t, m, commondao.ChoiceNo))
			}
			for _, m := range tc.abstain {
				record.AddVote(mustNewVote(t, m, commondao.ChoiceAbstain))
			}

			got := commondao.TallyDefault(record.Readonly(), electorate.Readonly(), tc.threshold)

			uassert.Equal(t, int(tc.want), int(got))
		})
	}
}

// refOutcome is an independent restatement of the constitutional tally
// rule, used to exhaustively cross-check TallyDefault at every boundary.
func refOutcome(electorateSize, yes, no, abstain int, threshold commondao.Threshold) commondao.Outcome {
	d := electorateSize - abstain
	if d <= 0 {
		return commondao.OutcomePending
	}
	passed := false
	if threshold == commondao.ThresholdSimpleMajority {
		passed = 2*yes > d // "more than half"
	} else {
		passed = 3*yes >= 2*d // "two thirds or more"
	}
	if passed {
		return commondao.OutcomePassed
	}
	if 2*no > d { // NO simple majority dismisses (both thresholds)
		return commondao.OutcomeDismissed
	}
	return commondao.OutcomePending
}

// TestTallyDefaultProperty exhaustively enumerates every electorate size
// up to 8 and every (yes, no, abstain) split for both thresholds,
// checking TallyDefault against refOutcome and machine-checking the
// invariants the ADR argues in prose (mutual exclusivity, the D>0 guard,
// abstain-shrinks-D monotonicity, and electorate gating).
func TestTallyDefaultProperty(t *testing.T) {
	// Eight electorate members plus one non-member "stranger".
	pool := []address{memberA, memberB, memberC, memberD, memberE, memberF, memberG, memberH}
	stranger := memberI
	thresholds := []commondao.Threshold{commondao.ThresholdSupermajority, commondao.ThresholdSimpleMajority}

	for e := 0; e <= len(pool); e++ {
		for yes := 0; yes <= e; yes++ {
			for no := 0; no+yes <= e; no++ {
				for abstain := 0; abstain+no+yes <= e; abstain++ {
					for _, th := range thresholds {
						var electorate addrset.Set
						for i := 0; i < e; i++ {
							electorate.Add(pool[i])
						}

						var record commondao.VotingRecord
						idx := 0
						put := func(n int, c commondao.VoteChoice) {
							for i := 0; i < n; i++ {
								v, err := commondao.NewVote(pool[idx], c, "")
								urequire.NoError(t, err)
								record.AddVote(v)
								idx++
							}
						}
						put(yes, commondao.ChoiceYes)
						put(no, commondao.ChoiceNo)
						put(abstain, commondao.ChoiceAbstain)
						// Remaining electorate members stay silent.

						got := commondao.TallyDefault(record.Readonly(), electorate.Readonly(), th)
						want := refOutcome(e, yes, no, abstain, th)
						urequire.Equal(t, int(want), int(got),
							"TallyDefault disagreed with the reference formula at some (E, yes, no, abstain, threshold)")

						// Invariant: pass and dismiss are mutually exclusive
						// (yes+no <= D holds since yes+no+abstain <= e).
						d := e - abstain
						passSuper := d > 0 && 3*yes >= 2*d
						passSimple := d > 0 && 2*yes > d
						dismiss := d > 0 && 2*no > d
						uassert.False(t, passSuper && dismiss, "supermajority pass and dismiss both fired")
						uassert.False(t, passSimple && dismiss, "simple pass and dismiss both fired")

						// Invariant: votes from outside the electorate never
						// change the outcome.
						sv, err := commondao.NewVote(stranger, commondao.ChoiceYes, "")
						urequire.NoError(t, err)
						record.AddVote(sv)
						gotWithStranger := commondao.TallyDefault(record.Readonly(), electorate.Readonly(), th)
						uassert.Equal(t, int(got), int(gotWithStranger), "out-of-electorate vote changed the outcome")
					}
				}
			}
		}
	}
}

// TestTallyDefaultDGuard pins the load-bearing D>0 guard: an all-abstain
// or empty electorate never passes, even at a zero-YES tally.
func TestTallyDefaultDGuard(t *testing.T) {
	var empty addrset.Set
	var record commondao.VotingRecord
	uassert.Equal(t, int(commondao.OutcomePending),
		int(commondao.TallyDefault(record.Readonly(), empty.Readonly(), commondao.ThresholdSupermajority)))

	var electorate addrset.Set
	electorate.Add(memberA)
	electorate.Add(memberB)
	v, err := commondao.NewVote(memberA, commondao.ChoiceAbstain, "")
	urequire.NoError(t, err)
	record.AddVote(v)
	v, err = commondao.NewVote(memberB, commondao.ChoiceAbstain, "")
	urequire.NoError(t, err)
	record.AddVote(v)
	// D = 2 - 2 = 0: no YES, must stay pending (not a 3*0 >= 2*0 pass).
	uassert.Equal(t, int(commondao.OutcomePending),
		int(commondao.TallyDefault(record.Readonly(), electorate.Readonly(), commondao.ThresholdSupermajority)))
}

// TestVotingRecordReVoteSameChoice pins that re-voting the identical choice
// is idempotent: AddVote decrements the previous choice then increments the
// new one, so a same-choice re-vote nets to zero — the counter must stay 1,
// not double-count.
func TestVotingRecordReVoteSameChoice(t *testing.T) {
	var record commondao.VotingRecord

	record.AddVote(mustNewVote(t, memberA, commondao.ChoiceYes))
	updated := record.AddVote(mustNewVote(t, memberA, commondao.ChoiceYes))

	uassert.True(t, updated, "expect a re-vote to be an update")
	uassert.Equal(t, 1, record.Size())
	uassert.Equal(t, 1, record.VoteCount(commondao.ChoiceYes))
}
