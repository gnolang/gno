package daocond

import (
	"errors"
	"math"

	"gno.land/p/nt/ufmt"
)

type membersThresholdCond struct {
	isMemberFn     func(memberId string) bool
	membersCountFn func() uint64
	threshold      float64
}

// Creates a condition requiring a percentage of members to vote yes.
//
// Example: MembersThreshold(0.6, isMemberFn, countFn)
// Requires 60% member approval.
func MembersThreshold(threshold float64, isMemberFn func(memberId string) bool, membersCountFn func() uint64) Condition {
	if threshold <= 0 || threshold > 1 {
		panic(errors.New("invalid threshold"))
	}
	if isMemberFn == nil {
		panic(errors.New("nil isMemberFn"))
	}
	if membersCountFn == nil {
		panic(errors.New("nil membersCountFn"))
	}
	return &membersThresholdCond{
		threshold:      threshold,
		isMemberFn:     isMemberFn,
		membersCountFn: membersCountFn,
	}
}

// Checks if enough members voted yes to meet the threshold.
func (c *membersThresholdCond) Eval(ballot Ballot) bool {
	return c.voteRatio(ballot, VoteYes) >= c.threshold
}

// Returns progress toward meeting the threshold between 0.0 and 1.0.
func (c *membersThresholdCond) Signal(ballot Ballot) float64 {
	return math.Min(c.voteRatio(ballot, VoteYes)/c.threshold, 1)
}

// Displays the condition as text.
// Example output: "60% of members"
func (c *membersThresholdCond) Render() string {
	return ufmt.Sprintf("%g%% of members", c.threshold*100)
}

// Displays the condition with current vote counts.
// Example output: "60% of members must vote yes\n\nYes: 3/5 = 60%"
func (c *membersThresholdCond) RenderWithVotes(ballot Ballot) string {
	s := ""
	s += ufmt.Sprintf("%g%% of members must vote yes\n\n", c.threshold*100)
	s += ufmt.Sprintf("Yes: %d/%d = %g%%\n\n", c.totalVote(ballot, VoteYes), c.membersCountFn(), c.voteRatio(ballot, VoteYes)*100)
	s += ufmt.Sprintf("No: %d/%d = %g%%\n\n", c.totalVote(ballot, VoteNo), c.membersCountFn(), c.voteRatio(ballot, VoteNo)*100)
	s += ufmt.Sprintf("Abstain: %d/%d = %g%%\n\n", c.totalVote(ballot, VoteAbstain), c.membersCountFn(), c.voteRatio(ballot, VoteAbstain)*100)
	return s
}

var _ Condition = (*membersThresholdCond)(nil)

func (c *membersThresholdCond) voteRatio(ballot Ballot, vote Vote) float64 {
	return float64(c.totalVote(ballot, vote)) / float64(c.membersCountFn())
}

func (c *membersThresholdCond) totalVote(ballot Ballot, vote Vote) uint64 {
	total := uint64(0)
	ballot.Iterate(func(voter string, v Vote) bool {
		if v == vote && c.isMemberFn(voter) {
			total += 1
		}
		return false
	})
	return total
}
